import AppKit
import Foundation
import UserNotifications

func environmentFlagEnabled(_ name: String) -> Bool {
	guard let value = ProcessInfo.processInfo.environment[name]?.trimmingCharacters(in: .whitespacesAndNewlines).lowercased() else {
		return false
	}
	return ["1", "true", "yes", "on"].contains(value)
}

let debugLogURL: URL? = {
	guard environmentFlagEnabled("TASKLIGHT_FOCUS_DEBUG") else { return nil }
	let cache = FileManager.default.urls(for: .cachesDirectory, in: .userDomainMask).first ?? URL(fileURLWithPath: NSTemporaryDirectory())
	let dir = cache.appendingPathComponent("tasklight", isDirectory: true)
	try? FileManager.default.createDirectory(
		at: dir,
		withIntermediateDirectories: true,
		attributes: [.posixPermissions: 0o700]
	)
	return dir.appendingPathComponent("native-helper.log")
}()

func debugLog(_ message: String) {
	guard let debugLogURL, let data = "\(Date()) \(message)\n".data(using: .utf8) else { return }
	let fileManager = FileManager.default
	let attributes = try? fileManager.attributesOfItem(atPath: debugLogURL.path)
	let size = (attributes?[.size] as? NSNumber)?.intValue ?? 0

	if size > 1_048_576 || !fileManager.fileExists(atPath: debugLogURL.path) {
		try? data.write(to: debugLogURL, options: .atomic)
	} else if let handle = try? FileHandle(forWritingTo: debugLogURL) {
		_ = try? handle.seekToEnd()
		try? handle.write(contentsOf: data)
		try? handle.close()
	}
	try? fileManager.setAttributes([.posixPermissions: 0o600], ofItemAtPath: debugLogURL.path)
}

func fail(_ message: String, code: Int32 = 1) -> Never {
	debugLog("fail: \(message)")
	fputs("TasklightNotifier: \(message)\n", stderr)
	exit(code)
}

struct Options {
	var title = "Tasklight"
	var subtitle = ""
	var message = ""
	var clickCommand = ""
	var sound = false
	var timeoutSeconds: TimeInterval = 60 * 60
}

final class NotificationDelegate: NSObject, UNUserNotificationCenterDelegate {
	var fallbackClickCommand: String
	var sound: Bool

	init(fallbackClickCommand: String = "", sound: Bool = false) {
		self.fallbackClickCommand = fallbackClickCommand
		self.sound = sound
	}

	func userNotificationCenter(
		_ center: UNUserNotificationCenter,
		willPresent notification: UNNotification,
		withCompletionHandler completionHandler: @escaping (UNNotificationPresentationOptions) -> Void
	) {
		if #available(macOS 11.0, *) {
			completionHandler(sound ? [.banner, .list, .sound] : [.banner, .list])
		} else {
			completionHandler(sound ? [.alert, .sound] : [.alert])
		}
	}

	func userNotificationCenter(
		_ center: UNUserNotificationCenter,
		didReceive response: UNNotificationResponse,
		withCompletionHandler completionHandler: @escaping () -> Void
	) {
		let userInfoCommand = response.notification.request.content.userInfo["clickCommand"] as? String ?? ""
		let command = userInfoCommand.isEmpty ? fallbackClickCommand : userInfoCommand
		debugLog("received notification response action=\(response.actionIdentifier) hasCommand=\(!command.isEmpty)")
		if response.actionIdentifier == UNNotificationDefaultActionIdentifier && !command.isEmpty {
			runShell(command)
		}
		completionHandler()
		DispatchQueue.main.asyncAfter(deadline: .now() + 0.2) {
			exit(0)
		}
	}
}

func parseOptions(_ args: [String]) throws -> Options {
	var options = Options()
	var index = 0

	if args.first == "notify" {
		index = 1
	}

	while index < args.count {
		let arg = args[index]
		switch arg {
		case "--title":
			options.title = try value(after: arg, args: args, index: &index)
		case "--subtitle":
			options.subtitle = try value(after: arg, args: args, index: &index)
		case "--message":
			options.message = try value(after: arg, args: args, index: &index)
		case "--click-command":
			options.clickCommand = try value(after: arg, args: args, index: &index)
		case "--timeout":
			let raw = try value(after: arg, args: args, index: &index)
			guard let timeout = TimeInterval(raw), timeout >= 0 else {
				throw UsageError("invalid --timeout: \(raw)")
			}
			options.timeoutSeconds = timeout
		case "--sound":
			options.sound = true
			index += 1
		case "-h", "--help", "help":
			printHelp()
			exit(0)
		default:
			throw UsageError("unexpected argument: \(arg)")
		}
	}

	if options.subtitle.isEmpty && options.message.isEmpty {
		throw UsageError("missing notification text; provide --subtitle or --message")
	}

	return options
}

func value(after flag: String, args: [String], index: inout Int) throws -> String {
	let valueIndex = index + 1
	guard valueIndex < args.count else {
		throw UsageError("missing value for \(flag)")
	}
	index += 2
	return args[valueIndex]
}

struct UsageError: Error, CustomStringConvertible {
	let description: String
	init(_ description: String) { self.description = description }
}

func printHelp() {
	print("""
	TasklightNotifier helper

	Usage:
	  TasklightNotifier notify --title <title> --subtitle <subtitle> --message <message> [--click-command <command>] [--sound] [--timeout <seconds>]
	  TasklightNotifier doctor [--timeout <seconds>]
	  TasklightNotifier self-test [--timeout <seconds>]
	""")
}

func internalTimeout(_ args: [String], default defaultValue: TimeInterval) throws -> TimeInterval {
	if args.count == 1 {
		return defaultValue
	}
	guard args.count == 3, args[1] == "--timeout", let timeout = TimeInterval(args[2]), timeout >= 0 else {
		throw UsageError("usage: \(args.first ?? "command") [--timeout <seconds>]")
	}
	return timeout
}

func runShell(_ command: String) {
	debugLog("running click command")
	let process = Process()
	process.executableURL = URL(fileURLWithPath: "/bin/sh")
	process.arguments = ["-c", command]
	do {
		try process.run()
	} catch {
		debugLog("failed to run click command: \(error)")
		fputs("TasklightNotifier: failed to run click command: \(error)\n", stderr)
	}
}

func requestAuthorization(center: UNUserNotificationCenter, sound: Bool) -> Bool {
	let semaphore = DispatchSemaphore(value: 0)
	var allowed = false
	var requestOptions: UNAuthorizationOptions = [.alert]
	if sound {
		requestOptions.insert(.sound)
	}

	center.requestAuthorization(options: requestOptions) { granted, error in
		if let error {
			debugLog("notification authorization error: \(error)")
			fputs("TasklightNotifier: notification authorization error: \(error)\n", stderr)
		}
		allowed = granted
		debugLog("notification authorization granted=\(granted)")
		semaphore.signal()
	}

	if semaphore.wait(timeout: .now() + 5) == .timedOut {
		debugLog("notification authorization request timed out")
		return false
	}
	return allowed
}

func sendNotification(center: UNUserNotificationCenter, options: Options) -> Bool {
	let content = UNMutableNotificationContent()
	content.title = options.title
	content.subtitle = options.subtitle
	content.body = options.message
	if options.sound {
		content.sound = .default
	}
	content.userInfo = ["clickCommand": options.clickCommand]

	let request = UNNotificationRequest(
		identifier: "tasklight-\(UUID().uuidString)",
		content: content,
		trigger: nil
	)

	let semaphore = DispatchSemaphore(value: 0)
	var ok: Bool?
	center.add(request) { error in
		if let error {
			debugLog("failed to add notification: \(error)")
			fputs("TasklightNotifier: failed to add notification: \(error)\n", stderr)
			ok = false
		} else {
			debugLog("notification added")
			ok = true
		}
		semaphore.signal()
	}

	if semaphore.wait(timeout: .now() + 5) == .timedOut {
		debugLog("adding notification timed out")
		return false
	}
	return ok ?? false
}

func notificationSettingName(_ setting: UNNotificationSetting) -> String {
	switch setting {
	case .notSupported: return "not-supported"
	case .disabled: return "disabled"
	case .enabled: return "enabled"
	@unknown default: return "unknown"
	}
}

func authorizationStatusName(_ status: UNAuthorizationStatus) -> String {
	switch status {
	case .notDetermined: return "not-determined"
	case .denied: return "denied"
	case .authorized: return "authorized"
	case .provisional: return "provisional"
	case .ephemeral: return "ephemeral"
	@unknown default: return "unknown"
	}
}

func printNotificationDiagnostics(timeoutSeconds: TimeInterval) {
	let center = UNUserNotificationCenter.current()
	let semaphore = DispatchSemaphore(value: 0)
	var settings: UNNotificationSettings?
	center.getNotificationSettings { currentSettings in
		settings = currentSettings
		semaphore.signal()
	}

	print("bundle-id=\(Bundle.main.bundleIdentifier ?? "missing")")
	if semaphore.wait(timeout: .now() + timeoutSeconds) == .timedOut || settings == nil {
		print("authorization=unavailable")
		print("alerts=unknown")
		print("sounds=unknown")
		return
	}
	guard let settings else { return }
	print("authorization=\(authorizationStatusName(settings.authorizationStatus))")
	print("alerts=\(notificationSettingName(settings.alertSetting))")
	print("sounds=\(notificationSettingName(settings.soundSetting))")
}

func runUntilClickOrTimeout(_ timeoutSeconds: TimeInterval) {
	let deadline = Date().addingTimeInterval(timeoutSeconds)
	debugLog("waiting for click until \(deadline)")
	while Date() < deadline {
		RunLoop.current.run(mode: .default, before: min(Date().addingTimeInterval(0.5), deadline))
	}
	debugLog("click wait timed out")
}

let rawArgs = Array(CommandLine.arguments.dropFirst())
debugLog("started argumentCount=\(rawArgs.count) bundle=\(Bundle.main.bundleIdentifier ?? "")")

if rawArgs.first == "doctor" {
	do {
		printNotificationDiagnostics(timeoutSeconds: try internalTimeout(rawArgs, default: 3))
		exit(0)
	} catch {
		fail("\(error)", code: 2)
	}
}

if rawArgs.first == "self-test" {
	do {
		runUntilClickOrTimeout(try internalTimeout(rawArgs, default: 0.05))
		print("timeout-exit=ok")
		exit(0)
	} catch {
		fail("\(error)", code: 2)
	}
}

NSApplication.shared.setActivationPolicy(.accessory)
let center = UNUserNotificationCenter.current()
let delegate = NotificationDelegate()
center.delegate = delegate

if rawArgs.isEmpty {
	debugLog("launched without notify args; waiting briefly for notification response")
	runUntilClickOrTimeout(30)
	debugLog("no-args response wait ended")
	exit(0)
}

let options: Options
do {
	options = try parseOptions(rawArgs)
	debugLog("parsed notification hasClick=\(!options.clickCommand.isEmpty) sound=\(options.sound)")
} catch {
	debugLog("usage error: \(error)")
	fputs("TasklightNotifier: \(error)\n", stderr)
	printHelp()
	exit(2)
}

delegate.fallbackClickCommand = options.clickCommand
delegate.sound = options.sound

if !requestAuthorization(center: center, sound: options.sound) {
	fail("notifications are not authorized")
}

if !sendNotification(center: center, options: options) {
	exit(1)
}

if options.clickCommand.isEmpty {
	debugLog("no click command; exiting after delivery grace period")
	RunLoop.current.run(until: Date().addingTimeInterval(1))
} else {
	debugLog("click command stored in notification userInfo; waiting for click")
	runUntilClickOrTimeout(options.timeoutSeconds)
}
debugLog("exiting")
