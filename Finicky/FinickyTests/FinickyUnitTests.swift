import Cocoa
@testable import Finicky
import XCTest

struct Fixture<T: Decodable & Equatable> {
    let value: T
    init(url: URL) throws {
        let actualData = try Data(contentsOf: url)
        value = try JSONDecoder().decode(T.self, from: actualData)
    }

    func assert(_ t: T) {
        XCTAssertEqual(t, value)
    }
}

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
        var capturedError: VersionParseError?
        let expected = [1, 2, 3, 4, 5, 6, 7]

        do {
            let actual = try makeVersionParts("1.2.3.4.5.6.7")
            XCTAssertEqual(actual, expected)
        } catch {
            capturedError = error as? VersionParseError
        }
        XCTAssertNil(capturedError)
    }

    func test_modificationDate() throws {
        let filePath = NSTemporaryDirectory().appending("test.json")
        let manager = FileManager()
        let now = Date()
        let createFile = manager.createFile(atPath: filePath, contents: Data("test".utf8), attributes: nil)
        precondition(createFile)
        let date = getModificationDate(fileManager: manager, atPath: filePath)
        XCTAssertEqual(0.1, date!.timeIntervalSince(now), accuracy: 0.1)
    }

    func test_getVersions() throws {
        let expected = Set([
            Version(title: "Finicky 2.2.3 - Minor fixes", version: "v2.2.3", prerelease: false),
            Version(title: "Finicky 2.3 ALPHA - Private/Incognito mode for Chrome", version: "v2.3-alpha", prerelease: true),
            Version(title: "Finicky 2.2.2 - Source process path & Dock icon bug fix", version: "v2.2.2", prerelease: false),
            Version(title: "Finicky 2.2.1 - Open preferred browsers", version: "v2.2.1", prerelease: false),
            Version(title: "Finicky 2.2.0 - Browser priority + block urls", version: "v2.2.0", prerelease: true),
            Version(title: "Finicky 2.1.0 - Keyboard support", version: "v2.1.0", prerelease: false),
            Version(title: "Finicky 2.0", version: "v2.0", prerelease: false),
            Version(title: "Finicky 2 RC3", version: "v2.0-rc.3", prerelease: true),
            Version(title: "Finicky 2 RC2", version: "v2.0-rc.2", prerelease: true),
            Version(title: "Finicky 2 RC1", version: "v2.0-rc.1", prerelease: true),
            Version(title: "Finicky 2 RC 0", version: "v2.0-rc.0", prerelease: true),
            Version(title: "Finicky v0.5", version: "v0.5", prerelease: true),
            Version(title: "Finicky v0.4", version: "v0.4", prerelease: true),
            Version(title: "Finicky v0.3", version: "v0.3", prerelease: true),
            Version(title: "Finicky v0.2", version: "v0.2", prerelease: true),
            Version(title: "Finicky v0.1", version: "v0.1", prerelease: true),
        ])

        let fixtureURL = Bundle(identifier: "net.kassett.FinickyTests")!
            .bundleURL
            .appendingPathComponent("Contents")
            .appendingPathComponent("Resources")
            .appendingPathComponent("Fixtures")
            .appendingPathComponent("github-fixture")
            .appendingPathExtension("json")

        let fixture = try Fixture<Set<Version>>(url: fixtureURL)
        fixture.assert(expected)

        let actualData = try Data(contentsOf: fixtureURL)
        let versions = try getVersions(data: actualData)

        XCTAssertEqual(versions, [
            Version(title: "Finicky 2.2.3 - Minor fixes", version: "v2.2.3", prerelease: false),
            Version(title: "Finicky 2.2.2 - Source process path & Dock icon bug fix", version: "v2.2.2", prerelease: false),
            Version(title: "Finicky 2.2.1 - Open preferred browsers", version: "v2.2.1", prerelease: false),
            Version(title: "Finicky 2.1.0 - Keyboard support", version: "v2.1.0", prerelease: false),
            Version(title: "Finicky 2.0", version: "v2.0", prerelease: false),
        ])
    }
}
