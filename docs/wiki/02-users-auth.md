# 02 · Người Dùng & Xác Thực

[← Wiki](./README.md)

---

## Người dùng (User)

### Các loại tài khoản

| Vai trò | Mô tả                                         |
| ------- | --------------------------------------------- |
| `user`  | Người dùng thông thường — tìm kiếm, đặt phòng |
| `admin` | Quản trị viên — toàn quyền quản lý hệ thống   |

### Trạng thái tài khoản

| Trạng thái | Ý nghĩa                            | Có thể đăng nhập?   |
| ---------- | ---------------------------------- | ------------------- |
| `active`   | Hoạt động bình thường              | ✅ Có               |
| `inactive` | Bị vô hiệu hóa thủ công            | ❌ Không            |
| `locked`   | Bị khóa do đăng nhập sai nhiều lần | ❌ Không (tạm thời) |

### Quy tắc thông tin cá nhân

- **Tên:** bắt buộc, 2–100 ký tự
- **Email:** bắt buộc, đúng định dạng, không được trùng trong hệ thống
- **Số điện thoại:** không bắt buộc, nếu nhập phải đúng 10 chữ số
- **Mật khẩu:** bắt buộc, tối thiểu 9 ký tự, tối đa 72 ký tự, phải có chữ hoa + chữ thường + số + ký tự đặc biệt

---

## Luồng đăng ký bằng email/mật khẩu

```
1. User nhập: tên, email, password, phone (tùy chọn)
2. Hệ thống tạo tài khoản (trạng thái: email chưa xác thực)
3. Hệ thống gửi email chứa link xác thực (hiệu lực 24 giờ)
4. User click link → email được xác thực
5. User có thể đăng nhập
```

> ⚠️ **Quan trọng:** Nếu chưa xác thực email, đăng nhập sẽ bị từ chối với lỗi "email chưa xác thực". User có thể yêu cầu gửi lại email xác thực bất cứ lúc nào.

---

## Luồng đăng nhập bằng email/mật khẩu

```
1. User nhập email + mật khẩu
2. Hệ thống kiểm tra theo thứ tự:
   a. Email có tồn tại không?
   b. Tài khoản có phải loại password không? (không phải OAuth)
   c. Tài khoản có đang bị khóa không?
   d. Mật khẩu có đúng không?
   e. Email đã xác thực chưa?
3. Nếu tất cả đúng → trả về Access Token + Refresh Token
```

### Bảo vệ tài khoản khỏi brute-force

Hệ thống tự động khóa tài khoản khi đăng nhập sai nhiều lần:

- **Ngưỡng kích hoạt:** 5 lần sai liên tiếp
- **Thời gian khóa tăng dần:**

| Lần bị khóa | Thời gian khóa        |
| ----------- | --------------------- |
| 1           | 15 phút               |
| 2           | 30 phút               |
| 3           | 1 giờ                 |
| 4           | 2 giờ                 |
| 5+          | tiếp tục tăng gấp đôi |

- Tài khoản **tự động mở khóa** sau khi hết thời gian — không cần admin can thiệp
- Đăng nhập thành công → reset về 0 lần sai
- Mức độ leo thang (LockEscalationLevel) **không reset** — lần khóa tiếp theo sẽ dài hơn

---

## Đăng nhập bằng Google (OAuth)

```
1. User click "Đăng nhập với Google"
2. Hệ thống redirect sang Google
3. User xác nhận trên Google
4. Google trả về thông tin user (email, tên, ảnh)
5. Hệ thống xử lý:
   - Nếu đã có account Google này → đăng nhập luôn
   - Nếu email đã tồn tại (đăng ký bằng password) → liên kết tài khoản
   - Nếu email mới → tạo tài khoản mới (email tự động xác thực)
6. Trả về Access Token + Refresh Token
```

### Quy tắc tài khoản OAuth

- Tài khoản tạo qua Google **không có mật khẩu**
- **Không thể đăng nhập** bằng email/password
- **Không thể reset mật khẩu** (vì không có mật khẩu)
- Email của tài khoản OAuth **tự động được xác thực** (do Google đã xác thực)

---

## Token xác thực

Hệ thống dùng 2 loại token:

| Token             | Thời hạn | Dùng để                                  |
| ----------------- | -------- | ---------------------------------------- |
| **Access Token**  | 60 phút  | Gọi các API được bảo vệ (gắn vào header) |
| **Refresh Token** | 30 ngày  | Lấy Access Token mới khi hết hạn         |

**Refresh Token rotation:** Mỗi lần refresh, token cũ bị vô hiệu hóa và token mới được cấp. Nếu phát hiện token cũ bị dùng lại → có thể đã bị đánh cắp.

**Revoke token khi:** đổi mật khẩu, reset mật khẩu, đăng xuất. Tất cả thiết bị đang đăng nhập sẽ bị đăng xuất.

---

## Reset mật khẩu

```
1. User nhập email → hệ thống gửi link reset (hiệu lực 30 phút)
2. User click link → nhập mật khẩu mới
3. Mật khẩu được cập nhật
4. Toàn bộ phiên đang đăng nhập bị đăng xuất (buộc đăng nhập lại)
```

> Lưu ý: Tài khoản OAuth (đăng ký qua Google) **không thể reset mật khẩu** vì không có mật khẩu.

---

## Các API liên quan

| Method | Endpoint                      | Mô tả                                   |
| ------ | ----------------------------- | --------------------------------------- |
| POST   | `/api/auth/register`          | Đăng ký                                 |
| POST   | `/api/auth/login`             | Đăng nhập                               |
| POST   | `/api/auth/logout`            | Đăng xuất _(yêu cầu JWT)_               |
| GET    | `/api/auth/me`                | Thông tin user hiện tại _(yêu cầu JWT)_ |
| POST   | `/api/auth/refresh`           | Làm mới token                           |
| POST   | `/api/auth/verify-email/send` | Gửi lại email xác thực                  |
| POST   | `/api/auth/verify-email`      | Xác thực email bằng token               |
| POST   | `/api/auth/forgot-password`   | Yêu cầu reset mật khẩu                  |
| POST   | `/api/auth/reset-password`    | Đặt mật khẩu mới                        |
| GET    | `/api/auth/google/login`      | Bắt đầu đăng nhập Google                |
| GET    | `/api/auth/google/callback`   | Callback từ Google                      |
| GET    | `/api/users`                  | Danh sách users                         |
| GET    | `/api/users/:id`              | Chi tiết user                           |
| POST   | `/api/users`                  | Tạo user mới                            |
| PATCH  | `/api/users/:id/profile`      | Cập nhật tên, SĐT                       |
| PATCH  | `/api/users/:id/password`     | Đổi mật khẩu                            |
| PATCH  | `/api/users/:id/deactivate`   | Vô hiệu hóa tài khoản                   |
