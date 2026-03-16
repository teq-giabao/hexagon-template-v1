# Wiki — Hệ Thống Đặt Phòng Khách Sạn

Tài liệu dành cho nhân viên mới và người maintain. Tập trung vào **nghiệp vụ** — cách hệ thống hoạt động, quy tắc kinh doanh, và các lưu ý quan trọng.

> Swagger UI (API chi tiết): `http://localhost:8088/swagger/index.html`

---

## Mục lục

| Tài liệu                                            | Nội dung                                    |
| --------------------------------------------------- | ------------------------------------------- |
| [01 · Tổng quan](./01-overview.md)                  | Hệ thống làm gì, ai dùng, tech stack        |
| [02 · Người dùng & Xác thực](./02-users-auth.md)    | Đăng ký, đăng nhập, OAuth, bảo vệ tài khoản |
| [03 · Khách sạn & Phòng](./03-hotel-room.md)        | Quản lý khách sạn, phòng, tồn kho           |
| [04 · Tìm kiếm & Đặt phòng](./04-search-booking.md) | Luồng tìm kiếm, quy tắc phân bổ phòng       |
| [05 · API Quick Reference](./05-api-reference.md)   | Danh sách endpoints, request mẫu            |

---

## Đọc nhanh — Những điều quan trọng nhất

Nếu chỉ có 5 phút, đọc những điều này trước:

1. **Email phải xác thực trước khi đăng nhập** — sau khi đăng ký, user nhận mail, click link xác nhận, sau đó mới login được.

2. **Tài khoản tự khóa sau 5 lần nhập sai mật khẩu** — thời gian khóa tăng dần (15 phút, 30 phút, 1 giờ...). Tự mở khóa sau khi hết thời gian.

3. **User đăng ký bằng Google không có mật khẩu** — không thể đăng nhập bằng email/password, không thể reset mật khẩu.

4. **Tồn kho phòng tính theo từng ngày** — mỗi phòng cần có inventory cho từng ngày cụ thể. Nếu thiếu ngày nào, phòng đó không hiện trong kết quả tìm kiếm.

5. **Trẻ em không được ở phòng không có người lớn** — đây là quy tắc nghiệp vụ quan trọng, được enforce trong toàn bộ luồng tìm kiếm và phân bổ phòng.
