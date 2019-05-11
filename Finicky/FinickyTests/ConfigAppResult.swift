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
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test Success")
        XCTAssertEqual(result!.appType, AppDescriptorType.appName)
    }

    func testObjectResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'Test Success' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test Success")
        XCTAssertEqual(result!.appType, AppDescriptorType.appName)
    }

    func testStringBundleIdResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "'test.success'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "test.success")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testObjectBundleIdResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "test.success")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testObjectFixedTypeResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success', appType: 'appName' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.appType, AppDescriptorType.appName)
    }

    func testObjectFixedTypeBundleIdResult() {
         _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success', appType: 'bundleId' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testObjectIncorrectObjectResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ wrong: 'test.success', bad: 'bundleId' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertNil(result)
    }

    func testObjectNoOpenInBackgroundBundleIdResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertNil(result!.openInBackground)
    }

    func testObjectOpenInBackgroundBundleIdResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success', openInBackground: true }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.openInBackground, true)
    }

    func testObjectDisableOpenInBackgroundBundleIdResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "{ name: 'test.success', openInBackground: false }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.openInBackground, false)
    }

    func testFunctionResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "(url, options) => { return { name: 'test.success', appType: 'bundleId' }}"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "test.success")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testFunctionOptionsResult() {
        _ = configLoader.parseConfig(generateHandlerConfig(app: "(url, options) => { return { name: options.url.protocol, appType: 'bundleId' }}"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "test.success")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }


}
