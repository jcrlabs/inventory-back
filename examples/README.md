# Examples

These JSON files are intended to be used with tools like curl, httpie, Postman or Insomnia.

Quick idea with curl:

```bash
export API_BASE="http://localhost:8080/api/v1"
TOKEN=$(curl -sS -X POST "$API_BASE/auth/login" -H 'Content-Type: application/json' -d @examples/login.json | jq -r .access_token)

curl -sS "$API_BASE/me" -H "Authorization: Bearer $TOKEN" | jq
curl -sS -X POST "$API_BASE/products" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d @examples/product_create.json | jq
```

Files:
- `login.json` – login body
- `admin_create_user.json` – admin create user body
- `admin_create_group.json` – admin create group body
- `admin_add_member.json` – admin adds group member body
- `product_create.json` – create product
- `product_update.json` – update product
- `add_image.json` – add image (after presign+upload)
- `contact.json` – upsert contact
