import JavaScriptCore

func generateConfig(defaultBrowser : String = "net.kassett.defaultBrowser", handlers: String) -> String {
    return """
        module.exports = {
            defaultBrowser: "\(defaultBrowser)",
            handlers: [\(handlers)]
        }
    """
}

func generateHandlerConfig(defaultBrowser : String = "'net.kassett.defaultBrowser'", match: String = "() => true", app: String = "'Test config'") -> String {

    try! evaluateScript(defaultBrowser)
    try! evaluateScript(match)
    try! evaluateScript(app)
    
    return """
        module.exports = {
            defaultBrowser: \(defaultBrowser),
            handlers: [{
                match: \(match),
                app: \(app)
            }]
        }
    """
}

enum ScriptEvaluationError: Error {
    case error(msg: String)
}

func evaluateScript(_ script: String) throws -> Void{
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
