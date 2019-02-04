import Cocoa
import XCTest
import Finicky
import JavaScriptCore

class FinickyTests: XCTestCase {

    var ctx : JSContext!
    var configLoader = FinickyConfig()
    let exampleUrl = URL(string: "http://example.com")!


    override func setUp() {
        super.setUp()
        _ = configLoader.createContext()
        FinickyAPI.reset()
        // Put setup code here. This method is called before the invocation of each test method in the class.
    }

    override func tearDown() {
        // Put teardown code here. This method is called after the invocation of each test method in the class.
        super.tearDown()
    }

    func generateConfig(defaultBrowser : String = "net.kassett.defaultBrowser", handlers: String) -> String {
        return """
            module.exports = {
                defaultBrowser: "\(defaultBrowser)",
                handlers: [\(handlers)]
            }
        """
    }

    func testStringMatcher() {
        _ = configLoader.parseConfig(generateConfig(handlers:
            """
                {
                    match: "http://example.com",
                    value: "Test Success"
                }
            """
        ))

        let result = configLoader.determineOpeningApp(url: URL(string: "http://example.com")!)
        XCTAssertEqual(result!.type, AppDescriptorType.appName,  "Type should be app name")
        XCTAssertEqual(result!.value, "Test Success", "App value should be set")
        XCTAssertEqual(result!.url.absoluteString, "http://example.com", "URL should not have been changed")
    }

    func testRegexMatcher() {
        _ = configLoader.parseConfig(generateConfig(handlers:
            """
                {
                    match: /http:\\/\\/example\\.com/,
                    value: "Test Success"
                }
            """
        ))

        let result = configLoader.determineOpeningApp(url: URL(string: "http://example.com")!)
        XCTAssertEqual(result!.type, AppDescriptorType.appName,  "Type should be app name")
        XCTAssertEqual(result!.value, "Test Success", "App value should be set")
        XCTAssertEqual(result!.url.absoluteString, "http://example.com", "URL should not have been changed")
    }

    func testFunctionMatcher() {
        _ = configLoader.parseConfig(generateConfig(handlers:
            """
                {
                    match: (url) => url === "http://example.com",
                    value: "Test Success"
                }
            """
        ))

        let result = configLoader.determineOpeningApp(url: URL(string: "http://example.com")!)
        XCTAssertEqual(result!.type, AppDescriptorType.appName,  "Type should be app name")
        XCTAssertEqual(result!.value, "Test Success", "App value should be set")
        XCTAssertEqual(result!.url.absoluteString, "http://example.com", "URL should not have been changed")
    }

    func testObjectResult() {
        _ = configLoader.parseConfig(generateConfig(handlers:
            """
                {
                    match: () => true,
                    value: { value: "Test Success", url: "http://another-example.com" }
                }
            """
        ))

        let result = configLoader.determineOpeningApp(url: URL(string: "http://example.com")!)
        XCTAssertEqual(result!.type, AppDescriptorType.appName,  "Type should be app name")
        XCTAssertEqual(result!.value, "Test Success", "App value should be set")
        XCTAssertEqual(result!.url.absoluteString, "http://example.com", "URL should not have been changed")
    }
}
