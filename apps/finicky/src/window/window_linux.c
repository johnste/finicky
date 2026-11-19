#include <gtk/gtk.h>
#include <webkit2/webkit2.h>
#include <stdlib.h>
#include <string.h>
#include "window_linux.h"
#include "_cgo_export.h"

// GTK may only be used from the thread running the main loop, so the
// functions called from Go schedule their work there with g_idle_add.

static GtkWidget *window = NULL;

typedef struct {
    char *url;
    char *userScript;
} ShowRequest;

// Hide rather than destroy on close, like the macOS window
// (setReleasedWhenClosed:NO), so it can be shown again.
static gboolean on_delete(GtkWidget *widget, GdkEvent *event, gpointer data) {
    gtk_widget_hide(widget);
    finickyWindowClosed();
    return TRUE;
}

// Links the user clicks, including ones opening a new window, are opened
// through Finicky's own rules instead of navigating the UI away.
static gboolean on_decide_policy(WebKitWebView *view, WebKitPolicyDecision *decision,
                                 WebKitPolicyDecisionType type, gpointer data) {
    if (type != WEBKIT_POLICY_DECISION_TYPE_NAVIGATION_ACTION &&
        type != WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION) {
        return FALSE;
    }
    WebKitNavigationAction *action = webkit_navigation_policy_decision_get_navigation_action(
        WEBKIT_NAVIGATION_POLICY_DECISION(decision));
    if (type == WEBKIT_POLICY_DECISION_TYPE_NEW_WINDOW_ACTION ||
        webkit_navigation_action_get_navigation_type(action) == WEBKIT_NAVIGATION_TYPE_LINK_CLICKED) {
        const char *uri = webkit_uri_request_get_uri(webkit_navigation_action_get_request(action));
        finickyOpenURL((char *)uri);
        webkit_policy_decision_ignore(decision);
        return TRUE;
    }
    return FALSE;
}

static GtkWidget *create_window(const char *url, const char *userScript) {
    GtkWidget *win = gtk_window_new(GTK_WINDOW_TOPLEVEL);
    gtk_window_set_title(GTK_WINDOW(win), "Finicky");
    gtk_window_set_default_size(GTK_WINDOW(win), 860, 600);
    gtk_widget_set_size_request(win, 800, 500);
    gtk_window_set_position(GTK_WINDOW(win), GTK_WIN_POS_CENTER);

    WebKitUserContentManager *contentManager = webkit_user_content_manager_new();
    WebKitUserScript *script = webkit_user_script_new(
        userScript, WEBKIT_USER_CONTENT_INJECT_TOP_FRAME,
        WEBKIT_USER_SCRIPT_INJECT_AT_DOCUMENT_START, NULL, NULL);
    webkit_user_content_manager_add_script(contentManager, script);
    webkit_user_script_unref(script);

    GtkWidget *view = webkit_web_view_new_with_user_content_manager(contentManager);
    webkit_settings_set_enable_developer_extras(
        webkit_web_view_get_settings(WEBKIT_WEB_VIEW(view)), TRUE);
    GdkRGBA background = {0.1, 0.1, 0.1, 1.0};
    webkit_web_view_set_background_color(WEBKIT_WEB_VIEW(view), &background);
    g_signal_connect(view, "decide-policy", G_CALLBACK(on_decide_policy), NULL);

    gtk_container_add(GTK_CONTAINER(win), view);
    g_signal_connect(win, "delete-event", G_CALLBACK(on_delete), NULL);
    webkit_web_view_load_uri(WEBKIT_WEB_VIEW(view), url);
    gtk_widget_show_all(view);
    return win;
}

static gboolean show_window_idle(gpointer data) {
    ShowRequest *req = data;
    if (!window) {
        window = create_window(req->url, req->userScript);
    }
    gtk_widget_show(window);
    gtk_window_present(GTK_WINDOW(window));
    free(req->url);
    free(req->userScript);
    free(req);
    return G_SOURCE_REMOVE;
}

static gboolean close_window_idle(gpointer data) {
    if (window && gtk_widget_get_visible(window)) {
        gtk_widget_hide(window);
        finickyWindowClosed();
    }
    return G_SOURCE_REMOVE;
}

void FinickyInitWindowing(void) {
    gtk_init(NULL, NULL);
}

void FinickyRunMainLoop(void) {
    gtk_main();
}

void FinickyShowWindow(const char *url, const char *userScript) {
    ShowRequest *req = malloc(sizeof(ShowRequest));
    req->url = strdup(url);
    req->userScript = strdup(userScript);
    g_idle_add(show_window_idle, req);
}

void FinickyCloseWindow(void) {
    g_idle_add(close_window_idle, NULL);
}
