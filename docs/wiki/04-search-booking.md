# 04 · Tìm Kiếm & Đặt Phòng

[← Wiki](./README.md)

---

## Tổng quan luồng tìm kiếm

Tìm kiếm diễn ra theo 3 bước, mỗi bước đào sâu hơn:

```
Bước 1  →  Tìm khách sạn phù hợp
Bước 2  →  Xem danh sách phòng available của khách sạn đó
Bước 3  →  Gợi ý tổ hợp phòng tối ưu để đặt
```

---

## Bước 1 · Tìm khách sạn

**Endpoint:** `POST /api/search/hotels`

### Tiêu chí tìm kiếm

| Tiêu chí                       | Bắt buộc | Mô tả                                      |
| ------------------------------ | -------- | ------------------------------------------ |
| `query`                        | ✅       | Từ khóa — tìm theo tên, thành phố, địa chỉ |
| `checkInAt`                    | ✅       | Ngày check-in (YYYY-MM-DD)                 |
| `checkOutAt`                   | ✅       | Ngày check-out (YYYY-MM-DD)                |
| `adultCount`                   | ✅       | Số người lớn (≥ 1)                         |
| `roomCount`                    | ✅       | Số phòng cần đặt (≥ 1)                     |
| `childrenAges`                 |          | Danh sách tuổi trẻ em, mỗi tuổi từ 0–17    |
| `ratingMin`                    |          | Đánh giá tối thiểu (0–5)                   |
| `amenityIds`                   |          | Lọc theo tiện nghi                         |
| `paymentOptions`               |          | Lọc theo phương thức thanh toán            |
| `page` / `pageSize` / `offset` |          | Phân trang (mặc định 10/trang)             |

### Quy tắc ngày

- Không được là ngày trong quá khứ
- CheckOut phải sau CheckIn
- Phải nằm trong cửa sổ đặt phòng hợp lệ

### Kết quả trả về

Mỗi khách sạn trong kết quả có 2 cờ quan trọng:

| Cờ                 | Ý nghĩa                                                 |
| ------------------ | ------------------------------------------------------- |
| `matchesRequested` | Có đủ phòng và đáp ứng được toàn bộ yêu cầu về số người |
| `flexibleMatch`    | Có phòng available nhưng không đủ để đáp ứng đầy đủ     |

Sắp xếp: `matchesRequested` hiện trước, rồi theo giá từ thấp đến cao.

---

## Bước 2 · Xem phòng của một khách sạn

**Endpoint:** `POST /api/search/hotels/:hotel_id/rooms`

Trả về danh sách phòng available của khách sạn trong khoảng ngày đó, kèm:

- Số lượng phòng còn trống (`availableCount`)
- Danh sách tiện nghi
- `strictMatch`: tổng phòng available ≥ số phòng yêu cầu

---

## Bước 3 · Gợi ý tổ hợp phòng

**Endpoint:** `POST /api/search/hotels/:hotel_id/room-combinations`

Đây là tính năng quan trọng nhất: **tìm cách ghép phòng tối ưu** để đủ chỗ cho toàn bộ nhóm khách, với chi phí thấp nhất.

**Ví dụ:** Nhóm 5 người lớn + 1 trẻ 5 tuổi, cần 2 phòng:

```
Phòng A: tối đa 2 người lớn + 1 trẻ, giá 1.000.000đ
Phòng B: tối đa 3 người lớn + 2 trẻ, giá 1.500.000đ

Tổ hợp gợi ý:
  1. 2×A = 4 người lớn + 2 trẻ, tổng 2.000.000đ  ✅
  2. 1×A + 1×B = 5 người lớn + 3 trẻ, tổng 2.500.000đ  ✅
  3. 2×B = 6 người lớn + 4 trẻ, tổng 3.000.000đ  ✅
```

Tham số `maxCombinations` (mặc định 5): giới hạn số tổ hợp trả về.

---

## Quy tắc phân bổ tuổi trẻ em

> ⚠️ **Đây là quy tắc nghiệp vụ quan trọng — cần nắm rõ trước khi sửa code liên quan.**

### Xác định trẻ em hay người lớn

Mỗi khách sạn có `DefaultChildMaxAge` (mặc định: **11**).

- Trẻ có tuổi **≤ DefaultChildMaxAge** → tính là **trẻ em**, chiếm slot MaxChild
- Trẻ có tuổi **> DefaultChildMaxAge** → tính là **người lớn**, chiếm slot MaxAdult

**Ví dụ** với `DefaultChildMaxAge = 11`:

- Trẻ 10 tuổi → trẻ em
- Trẻ 12 tuổi → người lớn (tốn slot người lớn)

### Quy tắc: Trẻ em không được ở một mình

**Mỗi phòng có trẻ em phải có ít nhất 1 người lớn đi kèm.**

Quy tắc này được kiểm tra **ở cả hai nơi:**

1. Khi tính `matchesRequested` trong kết quả tìm kiếm khách sạn
2. Khi generate tổ hợp phòng

Ví dụ vi phạm:

```
Yêu cầu: 0 người lớn, 2 trẻ em, 1 phòng
→ Bị từ chối — không thể để trẻ ở một mình
```

Ví dụ hợp lệ:

```
Yêu cầu: 1 người lớn, 2 trẻ em, 2 phòng
→ Phân bổ: Phòng 1: 1 người lớn + 1 trẻ ✅ | Phòng 2: 0 người lớn + 1 trẻ ❌
→ Hệ thống thử phân bổ khác hoặc báo không đủ phòng
```

---

## Lưu ý khi phát triển tính năng đặt phòng (chưa implement)

Khi implement booking trong tương lai, cần chú ý:

1. Khi tạo booking → tăng `HeldInventory` cho các ngày trong khoảng đặt phòng
2. Khi thanh toán thành công → giảm `HeldInventory`, tăng `BookedInventory`
3. Khi hủy booking → giảm `HeldInventory` hoặc `BookedInventory` tương ứng
4. Cần xử lý race condition khi nhiều user cùng đặt phòng cuối cùng

---

## Các API liên quan

| Method | Endpoint                                   | Mô tả                         |
| ------ | ------------------------------------------ | ----------------------------- |
| POST   | `/api/search/hotels`                       | Tìm khách sạn (có phân trang) |
| POST   | `/api/search/hotels/:id/rooms`             | Xem phòng available           |
| POST   | `/api/search/hotels/:id/room-combinations` | Gợi ý tổ hợp phòng            |
