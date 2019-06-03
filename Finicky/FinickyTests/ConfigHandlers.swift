import Cocoa
import Finicky
import JavaScriptCore
import XCTest

class ConfigHandlerTests: XCTestCase {
    var ctx: JSContext!
    var exampleUrl = URL(string: "http://example.com")!
    var configLoader: FinickyConfig!

    override func setUp() {
        super.setUp()
        configLoader = FinickyConfig()
        configLoader.setupAPI()
    }

    override func tearDown() {
        super.tearDown()
    }

    func testStringMatch() {
        configLoader.parseConfig(generateHandlerConfig(match: "'http://example.com'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testRegexMatcher() {
        configLoader.parseConfig(generateHandlerConfig(match: "/http:\\/\\/example\\.com/"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testFunctionMatcher() {
        configLoader.parseConfig(generateHandlerConfig(match: "(options) => options.urlString === 'http://example.com'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testFunctionMatcherDomainMatcher() {
        configLoader.parseConfig(generateHandlerConfig(match: "finicky.matchDomains(['example.com'])"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test config")
    }

    func testFunctionSourceBundleIdentifierMatcher() {
        configLoader.parseConfig(generateHandlerConfig(match:
            "(options) => options.sourceBundleIdentifier === 'testBundleId'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: "testBundleId")
        XCTAssertEqual(result!.name, "Test config")
    }

    func testDefaultBrowser() {
        configLoader.parseConfig(generateHandlerConfig(defaultBrowser: "'defaultBrowser'", match: "() => false"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "defaultBrowser")
    }

    func testInvalidConfig() {
        configLoader.parseConfig("!!! gibberish broken !!!")
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertNil(result)
    }
}
