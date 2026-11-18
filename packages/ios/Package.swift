// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "GotifyIOS",
    // macOS is listed so the package can also be built with just the Command
    // Line Tools on the Forgejo runner. A real .ipa still needs full Xcode.
    platforms: [.iOS(.v17), .macOS(.v13)],
    products: [
        .library(name: "GotifyIOS", targets: ["GotifyIOS"])
    ],
    targets: [
        .target(name: "GotifyIOS"),
        .testTarget(name: "GotifyIOSTests", dependencies: ["GotifyIOS"]),
    ]
)
