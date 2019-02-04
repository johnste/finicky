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

    fileprivate var shortUrlProviders = [
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
            if let responsy : HTTPURLResponse = response as? HTTPURLResponse {
                let newUrl = URL(string: (responsy.allHeaderFields["Location"] as? String)!)!
                callback(newUrl)
            } else {
                callback(url)
            }

        })

        task.resume()
        return
    }
}
