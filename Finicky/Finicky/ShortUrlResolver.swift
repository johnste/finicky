import Foundation

class ResolveShortUrls: NSObject, URLSessionDelegate, URLSessionTaskDelegate {

    fileprivate var shortUrlResolver : FNShortUrlResolver? = nil

    init(shortUrlResolver: FNShortUrlResolver) {
        self.shortUrlResolver = shortUrlResolver
        super.init()
    }

    func urlSession(_ session: URLSession, task: URLSessionTask, willPerformHTTPRedirection response: HTTPURLResponse, newRequest request: URLRequest, completionHandler: @escaping (URLRequest?) -> Void) {
        var newRequest : URLRequest? = request

        if [301, 302, 309].contains(response.statusCode) {
            if let newUrl = URL(string: (response.allHeaderFields["Location"] as? String)!) {
                if !shortUrlResolver!.isShortUrl(newUrl) {
                    newRequest = nil
                }
            }
        }

        completionHandler(newRequest)
    }
}

class FNShortUrlResolver {

    fileprivate var shortUrlProviders : [String] = []

    init(shortUrlProviders: [String]?) {
        self.shortUrlProviders = shortUrlProviders ?? [
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
            "tinyurl.com"
        ]
    }

    func isShortUrl(_ url: URL) -> Bool {
        return shortUrlProviders.contains((url.host!))
    }

    func resolveUrl(_ url: URL, callback: @escaping ((URL) -> Void)) -> Void {
        if !self.isShortUrl(url) {
            callback(url)
            return
        }

        let request = URLRequest(url: url)
        let myDelegate = ResolveShortUrls(shortUrlResolver: self)
        let session = URLSession(configuration: URLSessionConfiguration.default, delegate: myDelegate, delegateQueue: nil)

        let task = session.dataTask(with: request, completionHandler: { (data, response, error) -> Void in

            if let httpResponse : HTTPURLResponse = response as? HTTPURLResponse {
                let newUrl = URL(string: (httpResponse.allHeaderFields["Location"] as? String ?? url.absoluteString) )
                callback(newUrl ?? url)
            } else {
                callback(url)
            }

        })

        task.resume()
        return
    }
}
