#!/usr/bin/env python3
import asyncio
import websockets
import json

async def test_websocket_all():
    uri = "ws://127.0.0.1:8081/ws"
    headers = {
        "X-Node-ID": "xxx",
        "X-Secret": "xxx"
    }
    
    try:
        async with websockets.connect(uri, additional_headers=headers) as websocket:
            print("Connected to WebSocket")
            
            # 接收欢迎消息
            welcome = await websocket.recv()
            print(f"Server: {welcome}")
            
            # 测试 ping
            print("\n=== Testing Ping ===")
            ping_request = {
                "type": "ping",
                "payload": {
                    "host": "baidu.com",
                    "count": 2,
                    "timeout": 2
                }
            }
            print(f"Sending: {json.dumps(ping_request, indent=2)}")
            await websocket.send(json.dumps(ping_request))
            response = await websocket.recv()
            print(f"Response: {json.dumps(json.loads(response), indent=2)}")
            
            # 测试 tcping
            print("\n=== Testing TCPing ===")
            tcping_request = {
                "type": "tcping",
                "payload": {
                    "host": "baidu.com",
                    "port": 80,
                    "count": 2,
                    "timeout": 3
                }
            }
            print(f"Sending: {json.dumps(tcping_request, indent=2)}")
            await websocket.send(json.dumps(tcping_request))
            response = await websocket.recv()
            print(f"Response: {json.dumps(json.loads(response), indent=2)}")
            
            # 测试 website
            print("\n=== Testing Website ===")
            website_request = {
                "type": "website",
                "payload": {
                    "url": "https://www.baidu.com"
                }
            }
            print(f"Sending: {json.dumps(website_request, indent=2)}")
            await websocket.send(json.dumps(website_request))
            response = await websocket.recv()
            print(f"Response: {json.dumps(json.loads(response), indent=2)}")
            
            # 测试 dns
            print("\n=== Testing DNS ===")
            dns_request = {
                "type": "dns",
                "payload": {
                    "domain": "baidu.com",
                    "query_type": "A"
                }
            }
            print(f"Sending: {json.dumps(dns_request, indent=2)}")
            await websocket.send(json.dumps(dns_request))
            response = await websocket.recv()
            print(f"Response: {json.dumps(json.loads(response), indent=2)}")
            
            print("\n=== All tests completed ===")
            
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    asyncio.run(test_websocket_all())
