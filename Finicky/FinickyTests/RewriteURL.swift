import Cocoa
import XCTest
import Finicky
import JavaScriptCore

class URLRewriteTests: XCTestCase {

    var ctx : JSContext!
    var exampleUrl = URL(string: "http://example.com")!
    var configLoader : FinickyConfig!
    let urlMatch = "match: () => true"

    override func setUp() {
        super.setUp()
        configLoader = FinickyConfig()
        _ = configLoader.createContext()
    }

    override func tearDown() {
        super.tearDown()
    }

    func testRewrite() {
        _ = configLoader.parseConfig(generateRewriteConfig(url: "'http://test.com'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.url, URL(string: "http://test.com"))
    }

    func testRewriteFunction() {
        _ = configLoader.parseConfig(generateRewriteConfig(url: "() => 'http://test.com'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.url, URL(string: "http://test.com"))
    }

    func testRewriteFunctionArgs() {
        _ = configLoader.parseConfig(generateRewriteConfig(url: "(options) => options.urlString + '?ok'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.url, URL(string: "http://example.com?ok"))
    }

    func testRewriteFunctionArgs2() {
        _ = configLoader.parseConfig(generateRewriteConfig(url: "(options) => { finicky.log(options); return options.urlString + '?' + options.url.protocol; }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.url, URL(string: "http://example.com?http"))
    }
}
