//
//  NativeHost.swift
//  Finicky
//
//  Created by Alex Torma on 7/28/24.
//  Copyright © 2024 Alex Torma. All rights reserved.
//

import Foundation

func startListening() {
    // Function to read a message from stdin
    func readMessage() -> String? {
        let input = FileHandle.standardInput
        let data = input.availableData
        return String(data: data, encoding: .utf8)
    }

    // Function to write a message to stdout
    func writeMessage(_ message: String) {
        if let data = message.data(using: .utf8) {
            FileHandle.standardOutput.write(data)
        }
    }

    while true {
        if let message = readMessage() {
            // Process the message
            let response = "{\"text\": \"Received: \(message)\"}"
            writeMessage(response)
        }
    }
}
