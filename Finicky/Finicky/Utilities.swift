import Foundation

@discardableResult
func shell(_ args: String...) -> Int32 {
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

func getModificationDate(atPath: String) -> Date? {
    do {
        let fileManager = FileManager()
        let attributes = try fileManager.attributesOfItem(atPath: atPath)
        let modificationDate = attributes[FileAttributeKey.modificationDate] as? Date
        guard modificationDate != nil else { return nil }
        return modificationDate!
    } catch let msg {
        print("Error message: \(msg)")
        return nil
    }
}
