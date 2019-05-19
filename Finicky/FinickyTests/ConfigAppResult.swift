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
        configLoader.createContext()
    }

    override func tearDown() {
        super.tearDown()
    }

    func testStringResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "\"Test Success\""))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test Success")
        XCTAssertEqual(result!.appType, AppDescriptorType.appName)
    }

    func testObjectResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "{ name: 'Test Success' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "Test Success")
        XCTAssertEqual(result!.appType, AppDescriptorType.appName)
    }

    func testStringBundleIdResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "'test.success'"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "test.success")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testObjectBundleIdResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "{ name: 'test.success' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "test.success")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testObjectFixedTypeResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "{ name: 'test.success', appType: 'appName' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.appType, AppDescriptorType.appName)
    }

    func testObjectFixedTypeBundleIdResult() {
         configLoader.parseConfig(generateHandlerConfig(browser: "{ name: 'test.success', appType: 'bundleId' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testObjectIncorrectObjectResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "{ wrong: 'test.success', bad: 'bundleId' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertNil(result)
    }

    func testObjectNoOpenInBackgroundBundleIdResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "{ name: 'test.success' }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertNil(result!.openInBackground)
    }

    func testObjectOpenInBackgroundBundleIdResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "{ name: 'test.success', openInBackground: true }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.openInBackground, true)
    }

    func testObjectDisableOpenInBackgroundBundleIdResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "{ name: 'test.success', openInBackground: false }"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.openInBackground, false)
    }

    func testFunctionResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "(options) => { return { name: 'test.success', appType: 'bundleId' }}"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "test.success")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }

    func testFunctionOptionsResult() {
        configLoader.parseConfig(generateHandlerConfig(browser: "(options) => { return { name: options.url.protocol, appType: 'bundleId' }}"))
        let result = configLoader.determineOpeningApp(url: exampleUrl)
        XCTAssertEqual(result!.name, "http")
        XCTAssertEqual(result!.appType, AppDescriptorType.bundleId)
    }


}
