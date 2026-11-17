// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "GotifyIOS",
    platforms: [.iOS(.v17)],
    products: [
        .library(name: "GotifyIOS", targets: ["GotifyIOS"])
    ],
    targets: [
        .target(name: "GotifyIOS"),
        .testTarget(name: "GotifyIOSTests", dependencies: ["GotifyIOS"]),
    ]
)
