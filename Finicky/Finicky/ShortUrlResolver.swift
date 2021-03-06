import Foundation

/*

 Whenever finicky recieves a url, it tries to figure out if it redirects to a url shortener service.
 If that's the case, it will try to resolve the destination url. If it didn't, we would not be able
 to match the intended url with the configured patterns.
 */
final class ResolveShortUrls: NSObject, URLSessionDelegate, URLSessionTaskDelegate {
    fileprivate var shortUrlResolver: FNShortUrlResolver

    init(shortUrlResolver: FNShortUrlResolver) {
        self.shortUrlResolver = shortUrlResolver
        super.init()
    }

    func urlSession(_: URLSession, task _: URLSessionTask, willPerformHTTPRedirection response: HTTPURLResponse, newRequest request: URLRequest, completionHandler: @escaping (URLRequest?) -> Void) {
        var newRequest: URLRequest? = request

        if [301, 302, 309].contains(response.statusCode) {
            guard let urlString = ((response.allHeaderFields["Location"] as? String)!).addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) else {
                newRequest = nil
                return
            }

            if let newUrl = URL(string: urlString) {
                if !shortUrlResolver.isShortUrl(newUrl) {
                    newRequest = nil
                }
            }
        }

        completionHandler(newRequest)
    }
}

final class FNShortUrlResolver {
    private var shortUrlProviders: [String] = []
    var version: String

    convenience init() {
        self.init(shortUrlProviders: [])
    }

    init(shortUrlProviders: [String]) {
        self.shortUrlProviders = shortUrlProviders
        version = Bundle.main.infoDictionary!["CFBundleShortVersionString"] as! String
    }

    func isShortUrl(_ url: URL) -> Bool {
        guard let host = url.host,
              shortUrlProviders.contains(host) else { return false }

        // Can't load insecure cleartext HTTP
        // https://stackoverflow.com/questions/31254725/transport-security-has-blocked-a-cleartext-http
        return url.scheme == "https"
    }

    func resolveUrl(_ url: URL, callback: @escaping ((URL) -> Void)) {
        if !isShortUrl(url) {
            callback(url)
            return
        }

        var request = URLRequest(url: url)
        request.setValue("finicky/\(version)", forHTTPHeaderField: "User-Agent")
        let myDelegate = ResolveShortUrls(shortUrlResolver: self)
        let session = URLSession(configuration: URLSessionConfiguration.default, delegate: myDelegate, delegateQueue: nil)

        let task = session.dataTask(with: request, completionHandler: { (_, response, _) -> Void in
            if let httpResponse = response as? HTTPURLResponse {
                let urlString = (httpResponse.allHeaderFields["Location"] as? String ?? url.absoluteString).addingPercentEncoding(withAllowedCharacters: .urlQueryAllowed) ?? url.absoluteString

                let newUrl = URL(string: urlString)
                callback(newUrl ?? url)
            } else {
                callback(url)
            }
        })

        task.resume()
    }
}
