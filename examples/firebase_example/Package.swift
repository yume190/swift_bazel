// swift-tools-version: 5.7

import PackageDescription

let package = Package(
    name: "firebase_example",
    dependencies: [
        .package(url: "https://github.com/firebase/firebase-ios-sdk", .upToNextMajor(from: "10.0.0")),
    ]
)
