import CoreBluetooth
import CoreLocation
import Foundation
import Network

/// Reads the configured signals. Wi-Fi SSID is not available on iOS without a private
/// entitlement, so the iOS port uses geofences, Bluetooth and the manual override.
public final class PresenceDetectors: NSObject, CLLocationManagerDelegate {
    private let locationManager = CLLocationManager()
    private var lastLocation: CLLocation?
    private var connectedDeviceName: String?

    public var geofences: [(latitude: Double, longitude: Double, radius: CLLocationDistance)] = []
    public var bluetoothDeviceNames: Set<String> = []

    public override init() {
        super.init()
        locationManager.delegate = self
        locationManager.desiredAccuracy = kCLLocationAccuracyHundredMeters
    }

    public func requestPermissions() {
        locationManager.requestWhenInUseAuthorization()
    }

    public func startMonitoring() {
        for (index, fence) in geofences.enumerated() {
            let region = CLCircularRegion(
                center: CLLocationCoordinate2D(latitude: fence.latitude, longitude: fence.longitude),
                radius: fence.radius,
                identifier: "presence-\(index)"
            )
            region.notifyOnEntry = true
            region.notifyOnExit = true
            if CLLocationManager.isMonitoringAvailable(for: CLCircularRegion.self) {
                locationManager.startMonitoring(for: region)
            }
        }
        locationManager.startUpdatingLocation()
    }

    public func locationManager(_ manager: CLLocationManager, didUpdateLocations locations: [CLLocation]) {
        lastLocation = locations.last
    }

    public func locationManager(_ manager: CLLocationManager, didEnterRegion region: CLRegion) {
        lastLocation = manager.location
    }

    public func locationManager(_ manager: CLLocationManager, didExitRegion region: CLRegion) {
        lastLocation = manager.location
    }

    public func snapshot(manual: Signal) -> PresenceSnapshot {
        PresenceSnapshot(
            manual: manual,
            location: locationSignal(),
            bluetooth: Signal(
                configured: !bluetoothDeviceNames.isEmpty,
                available: !bluetoothDeviceNames.isEmpty,
                atDesk: connectedDeviceName != nil,
                label: connectedDeviceName
            )
        )
    }

    private func locationSignal() -> Signal {
        guard !geofences.isEmpty else { return Signal() }
        guard let location = lastLocation else { return Signal(configured: true) }
        let inside = geofences.first { fence in
            location.distance(
                from: CLLocation(latitude: fence.latitude, longitude: fence.longitude)
            ) <= fence.radius
        }
        return Signal(
            configured: true,
            available: true,
            atDesk: inside != nil,
            label: inside != nil ? "geofence" : nil
        )
    }
}
