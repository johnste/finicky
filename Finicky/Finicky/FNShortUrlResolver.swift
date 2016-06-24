//
//  ShortUrlResolver.swift
//  Finicky
//
//  Created by John Sterling on 03/07/15.
//  Copyright (c) 2015 John sterling. All rights reserved.
//

import Foundation

class ResolveShortUrls: NSObject, NSURLSessionDelegate, NSURLSessionTaskDelegate {

    private var shortUrlResolver : FNShortUrlResolver? = nil

    init(shortUrlResolver: FNShortUrlResolver) {
        self.shortUrlResolver = shortUrlResolver
        super.init()
    }

    func URLSession(session: NSURLSession, task: NSURLSessionTask, willPerformHTTPRedirection response: NSHTTPURLResponse, newRequest request: NSURLRequest, completionHandler: (NSURLRequest?) -> Void) {
        var newRequest : NSURLRequest? = request

        if [301, 302, 309].contains(response.statusCode) {
            if let newUrl = NSURL(string: (response.allHeaderFields["Location"] as? String)!) {
                if !shortUrlResolver!.isShortUrl(newUrl) {
                    newRequest = nil
                }
            }
        }

        completionHandler(newRequest)

    }
}

class FNShortUrlResolver {

    private var shortUrlProviders = [
        "bit.ly",
        "goo.gl",
        "ow.ly",
        "deck.ly",
        "t.co",
        "su.pr",
        "spoti.fi",
        "fur.ly",
        "tinyurl.com",
        "tiny.cc"
    ]

    init() {
    }

    func isShortUrl(url: NSURL) -> Bool {
        return shortUrlProviders.contains((url.host!))
    }

    func resolveUrl(url: NSURL, callback: ((NSURL) -> Void)) -> Void {
        if !self.isShortUrl(url) {
            callback(url)
            return
        }

        var response: NSURLResponse?
        var error: NSError?

        let request = NSURLRequest(URL: url)
        let myDelegate = ResolveShortUrls(shortUrlResolver: self)

        let session = NSURLSession(configuration: NSURLSessionConfiguration.defaultSessionConfiguration(), delegate: myDelegate, delegateQueue: nil)

        let task = session.dataTaskWithRequest(request, completionHandler: { (data, response, error) -> Void in
            if let responsy : NSHTTPURLResponse = response as? NSHTTPURLResponse {
                let newUrl = NSURL(string: (responsy.allHeaderFields["Location"] as? String)!)!
                callback(newUrl)
            } else {
                callback(url)
            }

        })


        task.resume()
        return
    }
}