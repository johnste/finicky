import Foundation

public struct Version: Decodable, Hashable {
    public let title: String
    public let version: String
    public let prerelease: Bool

    enum CodingKeys: String, CodingKey {
        case title = "name"
        case version = "tag_name"
        case prerelease
    }
}

enum defaultsKeys {
    static let keyLatestVersionSeen = "firstStringKey"
}

/*
 Check for available updates and call a callback if there is a new one available.
 Finicky saves the last seen version to avoid warning about version already seen.
 */

func checkForUpdate(_ alwaysNotify: Bool, _ newVersionCallback: @escaping (Version?, String) -> Void) {
    guard let currentVersion = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String,
          let url = URL(string: "https://api.github.com/repos/johnste/finicky/releases") else { return }
    var request = URLRequest(url: url)
    request.setValue("finicky/\(currentVersion)", forHTTPHeaderField: "User-Agent")

    let session = URLSession(configuration: URLSessionConfiguration.default, delegate: nil, delegateQueue: nil)
    let defaults = UserDefaults.standard

    let task = session.dataTask(with: request, completionHandler: { (data, _, _) -> Void in
        do {
            guard let data = data else { return }

            let versions = try getVersions(data: data)
            let sortedVersions = try versions.sorted(by: { (versionA, versionB) -> Bool in
                try compareVersions(versionA.version, versionB.version) == .orderedDescending
            })

            // Make sure we can get the latest version available
            guard let latestVersion = sortedVersions.first else { return }

            let latestSeenBefore = defaults.string(forKey: defaultsKeys.keyLatestVersionSeen)

            print("""
            Checking for updates:
                Current version: \(currentVersion)
                Available version: \(latestVersion.version)
                Latest version seen: \(latestSeenBefore ?? "<none>")
            """)

            let laterThanCurrent = try compareVersions(currentVersion, latestVersion.version)

            defaults.set(latestVersion.version, forKey: defaultsKeys.keyLatestVersionSeen)

            // We want to notify on versions we haven't seen before unless the user manually checks for updates
            if alwaysNotify || latestSeenBefore != latestVersion.version {
                if laterThanCurrent != ComparisonResult.orderedAscending {
                    // We running a newer or equal version to the available release
                    newVersionCallback(nil, currentVersion)
                    return
                } else {
                    newVersionCallback(latestVersion, currentVersion)
                }
            }
        } catch {
            print("error")
        }
    })

    task.resume()
}

// The `Error` protocol is used when throwing errors to catch
enum VersionParseError: Error {
    case badVersion(msg: String)
}

func makeVersionParts(_ version: String) throws -> [Int] {
    return try version.replacingOccurrences(of: #"[^0-9\.]"#, with: "", options: .regularExpression).split(separator: ".").map {
        guard let number = Int($0) else {
            throw VersionParseError.badVersion(msg: "Could not parse version: \(version)")
        }
        return number
    }
}

public func compareVersions(_ versionA: String, _ versionB: String) throws -> ComparisonResult {
    let partsA = try makeVersionParts(versionA)
    let partsB = try makeVersionParts(versionB)

    let seq = zip(partsA, partsB)

    for (el1, el2) in seq {
        if el1 < el2 {
            return .orderedAscending
        } else if el1 > el2 {
            return .orderedDescending
        }
    }

    if partsA.count < partsB.count {
        return .orderedAscending
    } else if partsA.count > partsB.count {
        return .orderedDescending
    }

    return .orderedSame
}

func getVersions(data: Data) throws -> Set<Version> {
    let versions = try JSONDecoder().decode(Set<Version>.self, from: data)
    return versions.filter { !$0.prerelease }
}
