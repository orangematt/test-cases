import ArgumentParser
import Foundation
import GRPC
import NIOCore
import SwiftProtobuf

class MessageClient : NSObject, ConnectivityStateDelegate {
    var group: EventLoopGroup?
    var client: Example_MessageServiceClient?
    var channel: GRPCChannel?

    var port = 9999
    var address = "localhost"

    deinit {
        disableRefresh()
    }

    func disableRefresh() {
        if let client = self.client {
            try? client.channel.close().wait()
            self.client = nil
            self.channel = nil
        } else if let channel = self.channel {
            try? channel.close().wait()
            self.channel = nil
        }
        if let group = self.group {
            try? group.syncShutdownGracefully()
            self.group = nil
        }
    }

    func enableRefresh() {
        if let _ = self.group {
            return
        }
        self.group = PlatformSupport.makeEventLoopGroup(loopCount: 1)
        if let group = self.group {
            let k = ClientConnectionKeepalive(interval: .minutes(2))
            let builder = ClientConnection.usingPlatformAppropriateTLS(for: group)
                .withConnectionReestablishment(enabled: true)
                .withConnectionTimeout(minimum: .seconds(2))
                .withKeepalive(k)
                .withConnectivityStateDelegate(self)
                .withConnectionIdleTimeout(.seconds(10))
            self.channel = builder.connect(host:self.address, port:self.port)
            self.streamUpdates()
        }
    }

    func streamUpdates() {
         if let channel = self.channel {
            self.client = Example_MessageServiceClient(channel:channel)
            if let client = self.client {
                let request = SwiftProtobuf.Google_Protobuf_Empty()
                print("Connection established. Sending initial request")
                let call = client.streamUpdates(request) { u in print("Incoming: \(u)") }

                let status = try! call.status.recover { _ in .processingError }.wait()
                if status.code != .ok {
                    print("RPC failed: \(status)")
                }
            }
        }
    }

    func connectivityStateDidChange(from oldState: ConnectivityState, to newState: ConnectivityState) {
        print("connectivityStateDidChange from", oldState, "to", newState)
        if oldState == .ready && newState == .idle {
            OperationQueue.main.addOperation() { self.streamUpdates() }
        }
    }

    func connectionStartedQuiescing() {
    }
}

struct ExampleClient: ParsableCommand {
    @Option var address: String?
    @Option var port: Int?

    mutating func run() throws {
        let c = MessageClient()
        if let address = address {
            c.address = address
        }
        if let port = port {
            c.port = port
        }
        c.enableRefresh()
    }
}

ExampleClient.main()
