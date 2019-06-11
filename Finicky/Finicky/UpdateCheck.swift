import Foundation

public struct Version {
    public var title: String
    public var version: String
    public var prerelease: Bool

    public init(title: String, version: String, prerelease: Bool) {
        self.title = title
        self.version = version
        self.prerelease = prerelease
    }
}

struct defaultsKeys {
    static let keyLatestVersionSeen = "firstStringKey"
}

func checkForUpdate(_ newVersionCallback: @escaping Callback<Version?>) {
    let version = Bundle.main.infoDictionary!["CFBundleShortVersionString"] as! String
    let url = URL(string: "https://api.github.com/repos/johnste/finicky/releases")
    guard url != nil else { return }
    var request = URLRequest(url: url!)
    request.setValue("finicky/\(version)", forHTTPHeaderField: "User-Agent")

    let session = URLSession(configuration: URLSessionConfiguration.default, delegate: nil, delegateQueue: nil)

    let defaults = UserDefaults.standard

    let task = session.dataTask(with: request, completionHandler: { (data, _, _) -> Void in
        do {
            if data == nil {
                return
            }

            let versions = try getVersions(data: data!)

            let sortedVersions = try versions.sorted(by: { (versionA, versionB) -> Bool in
                try compareVersions(versionA.version, versionB.version) == .orderedDescending
            })

            if let latestVersion = sortedVersions.first {
                let laterThanCurrent = try compareVersions(version, latestVersion.version)

                if laterThanCurrent == ComparisonResult.orderedAscending {
                    if let latestSeenBefore = defaults.string(forKey: defaultsKeys.keyLatestVersionSeen) {
                        print("latestSeenBefore \(latestSeenBefore)")
                        if latestSeenBefore != latestVersion.version {
                            defaults.set(latestVersion.version, forKey: defaultsKeys.keyLatestVersionSeen)
                            newVersionCallback(latestVersion)
                        }
                    }
                } else {
                    newVersionCallback(nil)
                }
            } else {
                newVersionCallback(nil)
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
    return try version.replacingOccurrences(of: #"[^0-9\.]"#, with: "", options: .regularExpression).split(separator: ".").map { (v) -> Int in
        let number = Int(v)

        guard number != nil else {
            throw VersionParseError.badVersion(msg: "Could not parse version: \(version)")
        }
        return number!
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

func getVersions(data: Data) throws -> [Version] {
    let json = try JSONSerialization.jsonObject(with: data) as! [AnyObject]

    let versions = json.map { (release: AnyObject?) -> Version? in
        if release == nil || release!["name"] == nil || release!["tag_name"] == nil || release!["prerelease"] == nil {
            return nil
        }

        return Version(title: release!["name"]! as! String, version: release!["tag_name"]! as! String, prerelease: release!["prerelease"] as! Bool)
    }.compactMap { $0 }.filter { !$0.prerelease }

    return versions
}
