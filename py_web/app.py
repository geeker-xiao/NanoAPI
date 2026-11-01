"""
NanoAPI 测试用 Flask 后端。

提供一组覆盖常见场景的接口，便于测试「请求 -> 响应」的各种返回形态：
  - 健康检查
  - 请求回显（echo）：把客户端发来的方法/参数/头/体原样返回，最便于排查
  - 列表（分页 + 过滤）
  - 详情（含 404 示例）
  - 创建（JSON 校验，含 400 示例）
  - 登录鉴权（含 401 示例）
  - 请求头查看

运行：
    cd py_web
    venv\\Scripts\\python.exe app.py
默认监听 http://127.0.0.1:5000

每个接口下方都附有「请求示例（curl）」与「响应示例（JSON）」。
"""

from __future__ import annotations

import json
import time
from typing import Any

from flask import Flask, jsonify, request

app = Flask(__name__)

# 关闭浏览器调用时的自动排序，保持字段顺序可读
app.config["JSON_SORT_KEYS"] = False
# 中文直接输出（不转义为 \uXXXX），便于阅读响应
app.json.ensure_ascii = False

# ---------------------------------------------------------------------------
# 示例数据（内存态，重启即重置）
# ---------------------------------------------------------------------------
USERS: list[dict[str, Any]] = [
    {"id": 1, "name": "张三", "email": "zhangsan@example.com", "role": "admin", "age": 28},
    {"id": 2, "name": "李四", "email": "lisi@example.com", "role": "user", "age": 34},
    {"id": 3, "name": "王五", "email": "wangwu@example.com", "role": "user", "age": 22},
    {"id": 4, "name": "赵六", "email": "zhaoliu@example.com", "role": "admin", "age": 41},
]
_NEXT_ID = 5

# 演示账号（仅用于示例）
DEMO_ACCOUNTS = {"admin": "123456", "user": "abcdef"}


# ---------------------------------------------------------------------------
# 通用工具
# ---------------------------------------------------------------------------
def success(data: Any = None, message: str = "ok", code: int = 0) -> tuple[Any, int]:
    """统一成功返回信封。"""
    return jsonify({"code": code, "message": message, "data": data}), 200


def error(message: str, http_status: int, code: int | None = None) -> tuple[Any, int]:
    """统一错误返回信封。"""
    return (
        jsonify({"code": code if code is not None else http_status, "message": message, "data": None}),
        http_status,
    )


@app.after_request
def _cors(resp):
    """允许跨域，方便用浏览器 / 前端 fetch 直接测试。"""
    resp.headers["Access-Control-Allow-Origin"] = "*"
    resp.headers["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
    resp.headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
    return resp


@app.errorhandler(404)
def _not_found(_e):
    return error("资源不存在 (Not Found)", 404)


@app.errorhandler(405)
def _method_not_allowed(_e):
    return error("方法不允许 (Method Not Allowed)", 405)


@app.errorhandler(500)
def _server_error(_e):
    return error("服务器内部错误 (Internal Server Error)", 500)


# ---------------------------------------------------------------------------
# 接口
# ---------------------------------------------------------------------------
@app.get("/")
def index():
    """
    接口目录 + 示例。

    响应示例：
    {
      "service": "NanoAPI Test Flask",
      "version": "1.0.0",
      "endpoints": ["/api/health", "/api/echo", "/api/users", ...]
    }
    """
    return success(
        {
            "service": "NanoAPI Test Flask",
            "version": "1.0.0",
            "endpoints": [
                "GET  /api/health",
                "GET  /api/echo",
                "GET  /api/headers",
                "GET  /api/users?page=1&limit=2&role=admin",
                "GET  /api/users/<id>",
                "POST /api/users",
                "POST /api/login",
            ],
        }
    )


@app.get("/api/health")
def health():
    """
    健康检查。

    请求示例：
        curl http://127.0.0.1:5000/api/health

    响应示例：
        {
          "code": 0,
          "message": "ok",
          "data": {"status": "healthy", "timestamp": 1762000000.123}
        }
    """
    return success({"status": "healthy", "timestamp": time.time()})


@app.route("/api/echo", methods=["GET", "POST", "PUT", "DELETE", "OPTIONS"])
def echo():
    """
    请求回显：把客户端发来的内容原样返回，最便于排查对接问题。

    请求示例（POST JSON）：
        curl -X POST http://127.0.0.1:5000/api/echo \\
             -H "Content-Type: application/json" \\
             -d '{"hello":"world","n":42}'

    响应示例：
        {
          "code": 0,
          "message": "ok",
          "data": {
            "method": "POST",
            "args": {},
            "headers": {"Host": "127.0.0.1:5000", "Content-Type": "application/json", ...},
            "json": {"hello": "world", "n": 42},
            "form": {},
            "raw": "{\"hello\":\"world\",\"n\":42}"
          }
        }
    """
    try:
        raw = request.get_data(as_text=True)
    except Exception:
        raw = ""
    return success(
        {
            "method": request.method,
            "path": request.path,
            "args": request.args.to_dict(),
            "headers": {k: v for k, v in request.headers.items()},
            "json": request.get_json(silent=True),
            "form": request.form.to_dict(),
            "raw": raw,
        }
    )


@app.get("/api/headers")
def headers():
    """
    仅返回请求头，便于确认客户端携带的 Header（如 Authorization）。

    请求示例：
        curl http://127.0.0.1:5000/api/headers \\
             -H "Authorization: Bearer demo-token-xxx" \\
             -H "X-Custom: abc"

    响应示例：
        {
          "code": 0,
          "message": "ok",
          "data": {"Authorization": "Bearer demo-token-xxx", "X-Custom": "abc", ...}
        }
    """
    return success({k: v for k, v in request.headers.items()})


@app.get("/api/users")
def list_users():
    """
    用户列表，支持分页与按 role 过滤。

    请求示例：
        curl "http://127.0.0.1:5000/api/users?page=1&limit=2&role=admin"

    响应示例：
        {
          "code": 0,
          "message": "ok",
          "data": {
            "items": [
              {"id": 1, "name": "张三", "role": "admin", "age": 28, "email": "..."},
              {"id": 4, "name": "赵六", "role": "admin", "age": 41, "email": "..."}
            ],
            "page": 1,
            "limit": 2,
            "total": 2
          }
        }
    """
    try:
        page = int(request.args.get("page", 1))
        limit = int(request.args.get("limit", 10))
    except ValueError:
        return error("page 与 limit 必须为整数", 400)

    role = request.args.get("role")
    items = USERS if not role else [u for u in USERS if u["role"] == role]

    start = max(page - 1, 0) * limit
    page_items = items[start : start + limit]
    return success({"items": page_items, "page": page, "limit": limit, "total": len(items)})


@app.get("/api/users/<int:user_id>")
def get_user(user_id: int):
    """
    用户详情。

    请求示例：
        curl http://127.0.0.1:5000/api/users/1

    成功响应示例：
        {"code": 0, "message": "ok",
         "data": {"id": 1, "name": "张三", "role": "admin", "age": 28, "email": "..."}}

    404 示例（id 不存在）：
        curl http://127.0.0.1:5000/api/users/999
        {"code": 404, "message": "资源不存在 (Not Found)", "data": null}
    """
    user = next((u for u in USERS if u["id"] == user_id), None)
    if user is None:
        return error("用户不存在", 404)
    return success(user)


@app.post("/api/users")
def create_user():
    """
    创建用户，演示 JSON 字段校验与 400 返回。

    请求示例（合法）：
        curl -X POST http://127.0.0.1:5000/api/users \\
             -H "Content-Type: application/json" \\
             -d '{"name":"钱七","email":"qianqi@example.com","role":"user","age":30}'

    成功响应示例（201）：
        {"code": 0, "message": "created",
         "data": {"id": 5, "name": "钱七", "role": "user", "age": 30, "email": "..."}}

    400 示例（缺 name / email 非法）：
        curl -X POST http://127.0.0.1:5000/api/users \\
             -H "Content-Type: application/json" -d '{"email":"bad"}'
        {"code": 400, "message": "字段校验失败",
         "data": {"name": "name 为必填字符串", "email": "email 格式不合法"}}
    """
    payload = request.get_json(silent=True)
    if not isinstance(payload, dict):
        return error("请求体必须是 JSON 对象", 400)

    errors: dict[str, str] = {}
    name = payload.get("name")
    email = payload.get("email")
    role = payload.get("role", "user")
    age = payload.get("age")

    if not isinstance(name, str) or not name.strip():
        errors["name"] = "name 为必填字符串"
    if not isinstance(email, str) or "@" not in email:
        errors["email"] = "email 格式不合法"
    if role not in ("admin", "user"):
        errors["role"] = "role 必须为 admin 或 user"
    if age is not None and (not isinstance(age, int) or age < 0 or age > 150):
        errors["age"] = "age 必须为 0-150 的整数"

    if errors:
        # 用统一信封返回，但把逐字段错误放进 data
        return (
            jsonify({"code": 400, "message": "字段校验失败", "data": errors}),
            400,
        )

    global _NEXT_ID
    new_user = {
        "id": _NEXT_ID,
        "name": name.strip(),
        "email": email,
        "role": role,
        "age": age,
    }
    USERS.append(new_user)
    _NEXT_ID += 1
    return jsonify({"code": 0, "message": "created", "data": new_user}), 201


@app.post("/api/login")
def login():
    """
    登录鉴权示例，演示 200 + token 与 401。

    请求示例（成功）：
        curl -X POST http://127.0.0.1:5000/api/login \\
             -H "Content-Type: application/json" \\
             -d '{"username":"admin","password":"123456"}'

    成功响应示例：
        {"code": 0, "message": "login success",
         "data": {"token": "demo-token-admin", "username": "admin", "role": "admin"}}

    401 示例（密码错误）：
        curl -X POST http://127.0.0.1:5000/api/login \\
             -H "Content-Type: application/json" -d '{"username":"admin","password":"wrong"}'
        {"code": 401, "message": "用户名或密码错误", "data": null}
    """
    payload = request.get_json(silent=True)
    if not isinstance(payload, dict):
        return error("请求体必须是 JSON 对象", 400)

    username = payload.get("username")
    password = payload.get("password")
    if not username or not password:
        return error("username 与 password 为必填", 400)

    expected = DEMO_ACCOUNTS.get(username)
    if expected is None or expected != password:
        return error("用户名或密码错误", 401)

    role = next((u["role"] for u in USERS if u["name"] == username), "user")
    return success(
        {"token": f"demo-token-{username}", "username": username, "role": role},
        message="login success",
    )


if __name__ == "__main__":
    # debug=True 便于改代码热重载；生产请关闭
    app.run(host="127.0.0.1", port=5000, debug=True)
