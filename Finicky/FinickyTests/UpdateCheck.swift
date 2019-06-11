import Cocoa
import Finicky
import XCTest
class CompareVersionTests: XCTestCase {
    override func setUp() {
        super.setUp()
    }

    override func tearDown() {
        super.tearDown()
    }

    func testRewrite() {
        XCTAssertEqual(try compareVersions("1.2.3", "0.9"), ComparisonResult.orderedAscending)
        XCTAssertEqual(try compareVersions("1.9.3", "0.9"), ComparisonResult.orderedAscending)
        XCTAssertEqual(try compareVersions("10", "9"), ComparisonResult.orderedAscending)
        XCTAssertEqual(try compareVersions("1.9.3.1", "1.9.3"), ComparisonResult.orderedAscending)
        XCTAssertEqual(try compareVersions("1.9.3", "1.9.3"), ComparisonResult.orderedAscending)
    }
}



