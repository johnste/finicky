import Cocoa
import XCTest
import Finicky
import JavaScriptCore

class ConfigHandlerTests: XCTestCase {

    var ctx : JSContext!
    var exampleUrl = URL(string: "http://example.com")!
    var configLoader : FinickyConfig!

    override func setUp() {
        super.setUp()
        configLoader = FinickyConfig()
        _ = configLoader.createContext()
    }

    override func tearDown() {
        super.tearDown()
    }
    func testStringMatch() {
        _ = configLoader.parseConfig(generateHandlerConfig(match: "'http://example.com'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testRegexMatcher() {
        _ = configLoader.parseConfig(generateHandlerConfig(match: "/http:\\/\\/example\\.com/"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testFunctionMatcher() {
        _ = configLoader.parseConfig(generateHandlerConfig(match: "(url) => url === 'http://example.com'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testFunctionMatcherDomainMatcher() {
        _ = configLoader.parseConfig(generateHandlerConfig(match: "finicky.matchDomains(['example.com'])"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testFunctionSourceBundleIdentifierMatcher() {
        _ = configLoader.parseConfig(generateHandlerConfig(match:
            "(url, options) => options.sourceBundleIdentifier === 'testBundleId'"
        ))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: "testBundleId")
        XCTAssertEqual(result!.name, "Test config")
    }

    func testDefaultBrowser() {
        _ = configLoader.parseConfig(generateHandlerConfig(defaultBrowser: "'defaultBrowser'", match: "() => false"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "defaultBrowser")
    }

    func testInvalidConfig() {
        _ = configLoader.parseConfig("!!! gibberish broken !!!")
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertNil(result)
    }

}
