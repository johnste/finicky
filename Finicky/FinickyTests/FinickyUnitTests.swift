import Cocoa
import XCTest
@testable import Finicky

class FinickyUnitTests: XCTestCase {
  func testRewrite() {
    XCTAssertEqual(try compareVersions("1.2.3", "0.9"), ComparisonResult.orderedDescending)
    XCTAssertEqual(try compareVersions("1.9.3", "0.9"), ComparisonResult.orderedDescending)
    XCTAssertEqual(try compareVersions("10", "9"), ComparisonResult.orderedDescending)
    XCTAssertEqual(try compareVersions("1.9.3.1", "1.9.3"), ComparisonResult.orderedDescending)
    XCTAssertEqual(try compareVersions("1.9.3", "1.9.3"), ComparisonResult.orderedSame)
    XCTAssertEqual(try compareVersions("0.9.3", "1.9.3"), ComparisonResult.orderedAscending)
  }

  func test_isShortUrl() {
    let resolver = FNShortUrlResolver(shortUrlProviders: ["te.st"])
    let actualPositive = resolver.isShortUrl(URL(string: "https://te.st/testing")!)
    let actualPositiveHttps = resolver.isShortUrl(URL(string: "http://te.st/testing")!)
    let actualNegative = resolver.isShortUrl(URL(string: "https://st.te/testing")!)

    XCTAssertTrue(actualPositive)
    XCTAssertFalse(actualNegative)
    XCTAssertFalse(actualPositiveHttps)
  }

  func test_notifications() throws {
    let center = NSUserNotificationCenter.default

    XCTAssertTrue(center.deliveredNotifications.isEmpty)
    showNotification(at: center, title: "title", subtitle: "subtitle", informativeText: "informativeText", error: true)

    XCTAssertFalse(center.deliveredNotifications.isEmpty)

    let notification = try XCTUnwrap(center.deliveredNotifications.first)

    XCTAssertEqual(notification.title, "title")
    XCTAssertEqual(notification.subtitle, "subtitle")
    XCTAssertEqual(notification.informativeText, "informativeText")
    XCTAssertEqual(notification.soundName, NSUserNotificationDefaultSoundName)

    center.removeDeliveredNotification(notification)
  }

  func test_makeVersionParts_error() {
    var capturedError: VersionParseError? = nil
    let expected = [1, 2, 3, 4, 5, 6, 7]

    do {
      let actual = try makeVersionParts("1.2.3.4.5.6.7")
      XCTAssertEqual(actual, expected)
    } catch {
      capturedError = error as? VersionParseError
    }
    XCTAssertNil(capturedError)
  }

}
