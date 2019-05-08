import Cocoa
import XCTest
import Finicky
import JavaScriptCore

class ConfigAppResultTests: XCTestCase {

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

    func testStringResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "\"Test Success\""))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: nil)
        XCTAssertEqual(result!.name, "Test Success", "App value should be set")
        XCTAssertEqual(result!.appType, AppDescriptorType.appName, "App type should be appName")
    }

    func testObjectResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'Test Success' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: nil)
        XCTAssertEqual(result!.name, "Test Success", "App value should be set")
        XCTAssertEqual(result!.appType, AppDescriptorType.appName, "App type should be appName")
    }

    func testStringBundleIdResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "\"test.success\""))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: nil)
        XCTAssertEqual(result!.name, "test.success", "App value should be set")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId, "App type should be bundle id")
    }

    func testObjectBundleIdResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: nil)
        XCTAssertEqual(result!.name, "test.success", "App value should be set")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId, "App type should be bundle id")
    }

    func testObjectFixedTypeResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success, appType: 'appName' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: nil)
        XCTAssertEqual(result!.appType, AppDescriptorType.appName, "App type should be bundle id")
    }

    func testObjectFixedTypeBundleIdResult() {
         _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success, appType: 'bundleId' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl, sourceBundleIdentifier: nil)
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId, "App type should be bundle id")
    }
}
