import Foundation

@discardableResult
func shell(_ args: [String]) -> Int32 {
    let task = Process()
    task.launchPath = "/usr/bin/env"
    task.arguments = args
    task.launch()
    task.waitUntilExit()
    return task.terminationStatus
}

func loadJS(_ path: String) -> String {
    let bundlePath = Bundle.main.path(forResource: path, ofType: nil)!
    return try! String(contentsOfFile: bundlePath, encoding: String.Encoding.utf8)
}

func getModificationDate(fileManager: FileManager = FileManager.default, atPath: String) -> Date? {
    do {
        let fileManager = FileManager()
        let attributes = try fileManager.attributesOfItem(atPath: atPath)
        return attributes[FileAttributeKey.modificationDate] as? Date
    } catch {
        print("Error message: \(error.localizedDescription)")
        return nil
    }
}

let defaultUrlShorteners = [
    "adf.ly",
    "bit.do",
    "bit.ly",
    "buff.ly",
    "deck.ly",
    "fur.ly",
    "goo.gl",
    "is.gd",
    "mcaf.ee",
    "ow.ly",
    "spoti.fi",
    "su.pr",
    "t.co",
    "tiny.cc",
    "tinyurl.com",
    "urlshortener.teams.microsoft.com",
]
