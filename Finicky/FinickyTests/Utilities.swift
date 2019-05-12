import JavaScriptCore

func generateConfig(defaultBrowser : String = "net.kassett.defaultBrowser", handlers: String) -> String {
    return """
        module.exports = {
            defaultBrowser: "\(defaultBrowser)",
            handlers: [\(handlers)]
        }
    """
}

func generateHandlerConfig(defaultBrowser : String = "'net.kassett.defaultBrowser'", match: String = "() => true", browser: String = "'Test config'") -> String {

    try! validateScript(defaultBrowser)
    try! validateScript(match)
    try! validateScript(browser)

    return """
        module.exports = {
            defaultBrowser: \(defaultBrowser),
            handlers: [{
                match: \(match),
                browser: \(browser)
            }]
        }
    """
}


func generateRewriteConfig(defaultBrowser : String = "'net.kassett.defaultBrowser'", match: String = "() => true", url: String = "'http://example.org'") -> String {

    try! validateScript(defaultBrowser)
    try! validateScript(match)
    try! validateScript(url)

    return """
    module.exports = {
        defaultBrowser: \(defaultBrowser),
        rewrite: [{
            match: \(match),
            url: \(url)
        }],
        handlers: []
    }
    """
}


enum ScriptEvaluationError: Error {
    case error(msg: String)
}

func validateScript(_ script: String) throws -> Void{
    let context = JSContext()!
    var error : String? = nil
    context.exceptionHandler = {
        context, exception in
        error = "\"" + String(describing: exception!) + "\" script: " + script
    }
    _ = context.evaluateScript("const finicky = { log() {}, matchDomains() {} }")
    _ = context.evaluateScript("const x = " + script)


    guard error == nil else {
        throw ScriptEvaluationError.error(msg: error!)
    }
}
