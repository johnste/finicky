
func generateConfig(defaultBrowser : String = "net.kassett.defaultBrowser", handlers: String) -> String {
    return """
        module.exports = {
            defaultBrowser: "\(defaultBrowser)",
            handlers: [\(handlers)]
        }
    """
}

func generateHandlerConfig(defaultBrowser : String = "net.kassett.defaultBrowser", match: String = "() => true", app: String = "\"Test config\"") -> String {
    return """
        module.exports = {
            defaultBrowser: "\(defaultBrowser)",
            handlers: [{
                match: \(match),
                app: \(app)
            }]
        }
    """
}
