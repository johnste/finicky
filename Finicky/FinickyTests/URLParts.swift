import Cocoa
import XCTest
import Finicky
class URLPartsTests: XCTestCase {

    override func setUp() {
        super.setUp()
    }

    override func tearDown() {
        super.tearDown()
    }

    func testRewrite() {
        let urlParts = FinickyAPI.getUrlParts("protocol://username:password@host:1234/pathname?search#hash")
        print(urlParts)
        XCTAssertEqual(urlParts["protocol"] as! String, "protocol")
        XCTAssertEqual(urlParts["username"] as! String, "username")
        XCTAssertEqual(urlParts["password"] as! String, "password")
        XCTAssertEqual(urlParts["host"] as! String, "host")
        XCTAssertEqual(urlParts["port"] as! Int, 1234)
        XCTAssertEqual(urlParts["pathname"] as! String, "/pathname")
        XCTAssertEqual(urlParts["search"] as! String, "search")
        XCTAssertEqual(urlParts["hash"] as! String, "hash")
    }
}
