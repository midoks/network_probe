#!/usr/bin/env python3
import asyncio
import websockets
import json

async def test_websocket_mtr():
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
            
            # 发送 mtr 测试请求
            mtr_request = {
                "type": "mtr",
                "payload": {
                    "host": "baidu.com",
                    "max_hops": 5,
                    "count": 3,
                    "interval": 1
                }
            }
            
            print(f"Sending: {json.dumps(mtr_request, indent=2)}")
            await websocket.send(json.dumps(mtr_request))
            
            # 接收响应
            response = await websocket.recv()
            print(f"Response: {json.dumps(json.loads(response), indent=2)}")
            
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    asyncio.run(test_websocket_mtr())
