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
class UpdateCheck {

    init() {

    }

    func check() {
        let version = Bundle.main.infoDictionary!["CFBundleShortVersionString"] as! String
        let url = URL(string: "https://api.github.com/repos/johnste/finicky/releases")
        guard url != nil else { return }
        var request = URLRequest(url: url!)
        request.setValue("finicky/\(version)", forHTTPHeaderField: "User-Agent")

        let session = URLSession(configuration: URLSessionConfiguration.default, delegate: nil, delegateQueue: nil)

        let task = session.dataTask(with: request, completionHandler: { (data, response, error) -> Void in
            do {
                let json = try JSONSerialization.jsonObject(with: data!) as! Array<AnyObject>

                let latestVersion = json.map({ (release: AnyObject?) -> Version? in
                    if( release == nil || release!["name"] == nil || release!["tag_name"] == nil || release!["prerelease"] == nil) {
                        return nil;
                    }

                    return Version(title: release!["name"]! as! String, version: release!["tag_name"]! as! String, prerelease: release!["prerelease"]! as! Bool)
                }).compactMap{ $0 }.filter({ $0.prerelease }).sorted(by: { $0.version > $1.version }).first

                print(latestVersion)
                if (latestVersion != nil) {
                    showNotification(title: "New version available: \(latestVersion!.title) \(latestVersion!.version)")
                }

            } catch {
                print("error")
            }
        })

        task.resume()
    }
}
