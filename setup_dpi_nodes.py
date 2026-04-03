import requests
import json
import subprocess
import time

BASE_URL = "http://localhost:4466/api"
EXTERNAL_IP = "95.85.116.86"

def clear_existing():
    print("Clearing existing inbounds...")
    resp = requests.get(f"{BASE_URL}/inbounds")
    inbounds = resp.json().get("obj", [])
    for ib in inbounds:
        requests.delete(f"{BASE_URL}/inbounds/{ib['id']}")
    print(f"Deleted {len(inbounds)} inbounds.")

def get_keygen(protocol):
    resp = requests.post(f"{BASE_URL}/clients/keygen/{protocol}")
    if resp.status_code == 200:
        return resp.json()["obj"]
    return None

def add_vless_reality():
    print("Adding VLESS REALITY (Port 993)...")
    keys = get_keygen("vless")
    if not keys:
        print("Failed to generate VLESS keys.")
        return

    priv = keys["private_key"]
    pub = keys["public_key"]
    short_id = "62a0df25"

    data = {
        "remark": "VLESS-REALITY-993",
        "protocol": "vless",
        "port": 993,
        "enable": True,
        "settings": json.dumps({
            "clients": [],
            "decryption": "none",
            "fallbacks": []
        }),
        "streamSettings": json.dumps({
            "network": "tcp",
            "security": "reality",
            "realitySettings": {
                "show": False,
                "dest": "belet.me:443",
                "xver": 0,
                "serverNames": ["belet.me"],
                "privateKey": priv,
                "publicKey": pub,
                "shortIds": [short_id]
            }
        }),
        "sniffing": json.dumps({"enabled": True, "destOverride": ["http", "tls"]})
    }
    resp = requests.post(f"{BASE_URL}/inbounds", json=data)
    ib_obj = resp.json().get("obj")
    if not ib_obj:
        print(f"Failed to add VLESS: {resp.text}")
        return
    ib_id = ib_obj["id"]
    
    # Add client
    requests.post(f"{BASE_URL}/inbounds/{ib_id}/clients", json={
        "email": "MainUser_VLESS",
        "id": "8bb8c8c8-c8c8-4c8c-bc8c-8c8c8c8c8c8c",
        "flow": "xtls-rprx-vision",
        "enable": True
    })
    print("VLESS REALITY added.")

def add_openvpn_xor():
    print("Adding OpenVPN XOR (Port 465)...")
    data = {
        "remark": "OpenVPN-XOR-465",
        "protocol": "openvpn",
        "port": 465,
        "enable": True,
        "settings": json.dumps({
            "proto": "udp",
            "cipher": "AES-256-GCM",
            "address": "10.8.0.0/24"
        }),
        "streamSettings": json.dumps({
            "network": "udp",
            "security": "none",
            "scramble_password": "snet_secure_xor"
        })
    }
    resp = requests.post(f"{BASE_URL}/inbounds", json=data)
    ib_obj = resp.json().get("obj")
    if not ib_obj:
        print(f"Failed to add OpenVPN: {resp.text}")
        return
    ib_id = ib_obj["id"]
    
    # Add client (OpenVPN XOR generates keys internally in adapter)
    requests.post(f"{BASE_URL}/inbounds/{ib_id}/clients", json={
        "email": "MainUser_OpenVPN",
        "enable": True
    })
    print("OpenVPN XOR added.")

def add_amnezia_wg():
    print("Adding AmneziaWG (Port 405)...")
    keys = get_keygen("amneziawg")
    if not keys:
        print("Failed to generate AmneziaWG keys.")
        return
    
    # AmneziaWG requires server-side keys in 'settings' JSON
    data = {
        "remark": "AmneziaWG-405",
        "protocol": "amneziawg",
        "port": 405,
        "enable": True,
        "settings": json.dumps({
            "address": "10.0.0.0/24",
            "private_key": keys["private_key"],
            "public_key": keys["public_key"]
        }),
        "streamSettings": json.dumps({
            "network": "udp",
            "security": "none",
            "S1": 15, "S2": 20, "H1": 1, "H2": 2, "H3": 3, "H4": 4
        })
    }
    resp = requests.post(f"{BASE_URL}/inbounds", json=data)
    ib_obj = resp.json().get("obj")
    if not ib_obj:
        print(f"Failed to add AmneziaWG: {resp.text}")
        return
    ib_id = ib_obj["id"]
    
    # Add client with its own keys
    c_keys = get_keygen("amneziawg")
    requests.post(f"{BASE_URL}/inbounds/{ib_id}/clients", json={
        "email": "MainUser_Amnezia",
        "enable": True,
        "privateKey": c_keys["private_key"],
        "publicKey": c_keys["public_key"],
        "presharedKey": c_keys.get("preshared_key", ""),
        "allowedIps": "10.0.0.2/32"
    })
    print("AmneziaWG added.")

if __name__ == "__main__":
    # Wait for backend to be ready if called sequentially
    time.sleep(5)
    clear_existing()
    add_vless_reality()
    add_openvpn_xor()
    add_amnezia_wg()
    print("\n--- Provisioning Complete ---")
    print(f"Server IP: {EXTERNAL_IP}")
    print("Ports: 993 (TCP), 465 (UDP), 405 (UDP)")
