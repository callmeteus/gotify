import SwiftUI

struct PresenceView: View {
    @State private var items: [Presence] = []
    @State private var message: String?

    private let baseURL: String
    private let token: String

    init(baseURL: String, token: String) {
        self.baseURL = baseURL
        self.token = token
    }

    var body: some View {
        List {
            if let message { Section { Text(message).foregroundStyle(.red) } }
            Section {
                Button("I'm at the PC") { override(state: .atDesk) }
                Button("I'm away") { override(state: .away) }
            }
            Section("Team") {
                ForEach(items, id: \.userId) { item in
                    HStack {
                        Text("User \(item.userId)")
                        Spacer()
                        Text(item.state.rawValue).foregroundStyle(.secondary)
                    }
                }
            }
        }
        .task { await load() }
    }

    private func endpoint(_ path: String) -> URL {
        var components = URLComponents(string: baseURL)!
        components.path += "/presence" + path
        components.queryItems = [URLQueryItem(name: "token", value: token)]
        return components.url!
    }

    private func load() async {
        do {
            let (data, _) = try await URLSession.shared.data(from: endpoint(""))
            items = try JSONDecoder().decode([Presence].self, from: data)
        } catch {
            message = "Could not load presence"
        }
    }

    private func override(state: PresenceState) {
        var request = URLRequest(url: endpoint(""))
        request.httpMethod = "PUT"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try? JSONSerialization.data(withJSONObject: [
            "state": state.rawValue,
            "source": "manual",
            "validForSeconds": 1800,
        ])
        Task {
            do {
                let (_, response) = try await URLSession.shared.data(for: request)
                guard (response as? HTTPURLResponse)?.statusCode == 200 else { return }
                await load()
            } catch {
                message = "Could not publish presence"
            }
        }
    }
}
