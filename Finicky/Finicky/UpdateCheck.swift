import Foundation

public struct Version: Decodable, Hashable  {
  public let title: String
  public let version: String
  public let prerelease: Bool

  enum CodingKeys: String, CodingKey {
    case title = "name"
    case version = "tag_name"
    case prerelease
  }
}

struct defaultsKeys {
    static let keyLatestVersionSeen = "firstStringKey"
}

func checkForUpdate(_ newVersionCallback: @escaping Callback<Version?>) {
    guard let version = Bundle.main.infoDictionary?["CFBundleShortVersionString"] as? String,
      let url = URL(string: "https://api.github.com/repos/johnste/finicky/releases") else {return}
    var request = URLRequest(url: url)
    request.setValue("finicky/\(version)", forHTTPHeaderField: "User-Agent")

    let session = URLSession(configuration: URLSessionConfiguration.default, delegate: nil, delegateQueue: nil)

    let defaults = UserDefaults.standard

    let task = session.dataTask(with: request, completionHandler: { (data, _, _) -> Void in
        do {
            guard let data = data else { return }

            let versions = try getVersions(data: data)
            let sortedVersions = try versions.sorted(by: { (versionA, versionB) -> Bool in
                try compareVersions(versionA.version, versionB.version) == .orderedDescending
            })

          guard let latestVersion = sortedVersions.first else { newVersionCallback(nil); return }
          let laterThanCurrent = try compareVersions(version, latestVersion.version)
          guard laterThanCurrent == ComparisonResult.orderedAscending else { newVersionCallback(nil); return }
          guard let latestSeenBefore = defaults.string(forKey: defaultsKeys.keyLatestVersionSeen) else { return }
            print("latestSeenBefore \(latestSeenBefore)")
            if latestSeenBefore != latestVersion.version {
              defaults.set(latestVersion.version, forKey: defaultsKeys.keyLatestVersionSeen)
              newVersionCallback(latestVersion)
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
  return versions.filter{!$0.prerelease}
}
