import Foundation

/// Connects to `/stream?presence=1`. Message frames are forwarded as raw JSON;
/// presence frames are decoded into `Presence`.
final class GotifyWebSocket: NSObject, URLSessionWebSocketDelegate {
    var onMessage: ((String) -> Void)?
    var onPresence: ((Presence) -> Void)?
    var onOpen: (() -> Void)?
    var onClose: (() -> Void)?

    private let url: URL
    private var task: URLSessionWebSocketTask?
    private let decoder: JSONDecoder = {
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        return decoder
    }()

    init(baseURL: String, token: String) {
        var components = URLComponents(string: baseURL)!
        components.path += "/stream"
        components.queryItems = [
            URLQueryItem(name: "token", value: token),
            URLQueryItem(name: "presence", value: "1"),
        ]
        url = components.url!
        super.init()
    }

    func connect() {
        task = URLSession.shared.webSocketTask(with: url)
        task?.delegate = self
        task?.resume()
        onOpen?()
        receive()
    }

    func close() {
        task?.cancel(with: .goingAway, reason: nil)
        task = nil
    }

    func urlSession(_ session: URLSession, webSocketTask: URLSessionWebSocketTask,
                           didCloseWith closeCode: URLSessionWebSocketTask.CloseCode, reason: Data?) {
        onClose?()
    }

    private func receive() {
        task?.receive { [weak self] result in
            guard let self else { return }
            switch result {
            case .success(.string(let text)):
                if let frame = try? self.decoder.decode(PresenceFrame.self, from: Data(text.utf8)),
                   frame.isPresence, let presence = frame.data {
                    self.onPresence?(presence)
                } else {
                    self.onMessage?(text)
                }
                self.receive()
            case .success:
                self.receive()
            case .failure:
                self.onClose?()
            }
        }
    }
}
