# URL Shortener Service

**[Xem bản tiếng Anh](README.md)**

Dịch vụ rút gọn link được xây dựng bằng Golang và React, cho phép người dùng tạo các link ngắn gọn, dễ nhớ kèm theo tính năng theo dõi lượt nhấp và quản lý thời hạn hết hạn.
Demo: https://shorty-black.vercel.app/home

---

## 📋 Table of Contents

- [Problem Description](#-problem-description)
- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Getting Started](#-getting-started)
- [Design & Technical Decisions](#-design--technical-decisions)
- [Security Considerations](#-security-considerations)
- [Scalability](#-scalability)
- [Trade-offs](#-trade-offs)
- [Challenges & Learnings](#-challenges--learnings)

---

## 🎯 Problem Description

Xây dựng dịch vụ rút gọn URL cho phép:
- Người dùng nhập URL dài → nhận lại một URL ngắn dễ nhớ.
- Truy cập URL ngắn → chuyển hướng về URL gốc.
- Theo dõi số lượng lượt nhấp chuột.
- Quản lý các link đã tạo kèm theo ngày hết hạn.

**Tóm tắt**: Tạo link ngắn, chuyển hướng chính xác, theo dõi lượt nhấp, ngăn chặn trùng lặp và quản lý link theo từng người dùng.

---

## ✨ Features

### Tính năng cốt lõi
- Tạo URL ngắn từ URL dài.
- Tự động chuyển hướng về URL gốc.
- Theo dõi lượt nhấp .
- Danh sách tất cả các link đã tạo.
- Quản lý thời gian hết hạn của link.
- Ngăn chặn việc rút gọn trùng lặp URL.

### Tính năng bổ sung
- Xác thực người dùng bằng JWT.
- Hỗ trợ tùy chỉnh định danh (custom alias).
- Tạo mã QR (thông qua Cloudinary).
- Giới hạn (100 URL mỗi người dùng/ngày).
- Kiểm tra tính hợp lệ của URL.

---

## 🛠 Tech Stack

### Backend
- **Ngôn ngữ**: Golang
- **Framework**: Gin
- **Cơ sở dữ liệu**: PostgreSQL (lưu trữ trên Neon)
- **Xác thực**: JWT
- **Lưu trữ tệp tin**: Cloudinary
- **Triển khai**: Fly.io

### Frontend
- **Framework**: React

---

## 🚀 Getting Started
```bash
# Clone dự án
git clone https://github.com/nhatcn/shorty.git
cd shorty
```

### Backend Setup

```bash
# Di chuyển vào thư mục backend
cd backend

# Tải các thư viện phụ thuộc
go mod download

# Tạo tệp .env
New-Item -Path . -Name ".env" -ItemType "File"
```

**Cấu hình `.env`:**
```env
DATABASE_URL=postgresql://neondb_owner:npg_xKsv3fSC5myF@ep-proud-shadow-a1gzrzfn-pooler.ap-southeast-1.aws.neon.tech/neondb?sslmode=require&channel_binding=require  (để tiện dùng thử)
JWT_SECRET=your_secret_key
PORT=8080
FRONTEND_URL=http://localhost:3000
CLOUDINARY_CLOUD_NAME=your_cloudinary_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

**Khởi chạy máy chủ backend:**
```bash
go run cmd/server/main.go
```
Backend chạy tại: `http://localhost:8080`

### Frontend Setup

```bash
# Di chuyển vào thư mục frontend
cd frontend

# Cài đặt thư viện
npm install

# Tạo tệp .env
New-Item -Path . -Name ".env" -ItemType "File"
```

**Cấu hình `.env`:**
```env
BE_URL=http://localhost:8080
```

**Khởi chạy môi trường phát triển:**
```bash
npm start
```
Frontend chạy tại: `http://localhost:3000`

---

## 🧠 Design & Technical Decisions

### Tại sao chọn PostgreSQL?
- Hỗ trợ SQL chuẩn, mạnh mẽ cho các truy vấn phức tạp (joins, aggregation, indexes).
- Giao dịch (Transactions) xử lý đồng thời an toàn (ví dụ: tạo shortCode từ ID tự tăng).
- Dễ dàng triển khai trên Fly.io, Neon, Railway với gói miễn phí và thiết lập nhanh.
- Hỗ trợ các ràng buộc và chỉ mục: UNIQUE, composite index, giúp tối ưu hiệu năng.

### Tại sao chọn RESTful API?
- Phổ biến, dễ hiểu, dễ kiểm thử với Postman/curl.
- Các endpoint rõ ràng: `POST /urls` để tạo, `GET /urls/:shortCode` để chuyển hướng.
- Dễ dàng mở rộng và tích hợp với các frontend như React/Vue.

### Thuật toán tạo mã rút gọn
**Phương pháp**: Sử dụng ID tự tăng của PostgreSQL → mã hóa sang Base62.
- ID tự tăng đảm bảo tính duy nhất 100%.
- Mã hóa ID sang Base62 để tạo ra shortCode ngắn gọn.

### Xử lý xung đột và trùng lặp
- **Trùng URL + Người dùng**: Nếu cùng một người dùng gửi lại URL cũ → trả về shortCode đã tồn tại.
- **Xung đột mã rút gọn**: Không bao giờ xảy ra nhờ sử dụng ID tự tăng mã hóa Base62.
- **Xử lý đồng thời**: Cơ sở dữ liệu xử lý các giao dịch đảm bảo không trùng lặp và tránh tình trạng race conditions.

---

## 🔒 Security Considerations

1. **Tính dự đoán được của mã rút gọn**: Có thể sử dụng các phương pháp làm xáo trộn ID (XOR, hash) để tránh lộ quy luật tạo link.
2. **Kiểm tra URL**: Ngăn chặn các địa chỉ nội bộ (localhost, IP riêng) và chỉ cho phép giao thức HTTP/HTTPS.
3. **Theo dõi lượt nhấp**: Kiểm tra tính hợp lệ của ID để ngăn chặn tấn công injection.
4. **Giới hạn tốc độ (Rate Limiting)**: Giới hạn 100 URL/ngày để ngăn chặn thư rác.
5. **URL hết hạn**: Kiểm tra `ExpiresAt` trước khi chuyển hướng.
6. **Xác thực và phân quyền**: Người dùng chỉ có thể xem hoặc xóa các URL do chính họ tạo ra.

---

## 📈 Scalability

### Xử lý khi lưu lượng truy cập tăng 100 lần

#### Ưu tiên Đọc (Read-Heavy)
- Sử dụng **Redis** để lưu trữ bộ nhớ đệm cho các ánh xạ URL (shortCode → originalURL) nhằm giảm tải cho DB.

#### Ưu tiên Ghi (Write-Heavy)
- Nếu có nhiều máy chủ, DB phải xử lý đồng thời tốt cho ID tự tăng. Lượt nhấp có thể được lưu theo lô (batch insert) để giảm số lượng lệnh ghi liên tục.

#### Phân mảnh cơ sở dữ liệu (Sharding/Partitioning)
- Khi dữ liệu vượt quá hàng trăm triệu dòng, có thể phân chia bảng theo `user_id` hoặc theo khoảng thời gian.

---

## ⚖️ Trade-offs

- **PostgreSQL vs NoSQL**: Chọn PostgreSQL vì giao dịch an toàn và các ràng buộc mạnh mẽ, dù việc mở rộng quy mô cực lớn có thể phức tạp hơn NoSQL.
- **REST vs GraphQL**: Chọn REST vì sự đơn giản và tốc độ phát triển nhanh cho dự án quy mô vừa và nhỏ.
- **ID tự tăng + Base62**: Đảm bảo tính duy nhất tuyệt đối nhưng mã rút gọn sẽ tăng dần theo tuần tự, có thể làm lộ thứ tự tạo link.

---

## 💡 Challenges & Learnings

### Các vấn đề gặp phải
- Lần đầu tiếp cận với Golang, còn lạ lẫm với cú pháp và cấu trúc package.
- Gặp lỗi khi kết nối và triển khai PostgreSQL trực tiếp trên Fly.io.

### Giải pháp
- Tự học cấu trúc Golang (handler → service → repository).
- Chuyển sang sử dụng cơ sở dữ liệu đám mây Neon (PostgreSQL cloud) để đảm bảo độ ổn định và triển khai nhanh.

### Bài học rút ra
- Hiểu rõ luồng hoạt động của Backend trong Golang.
- Kỹ năng xử lý đồng thời, trùng lặp và giao dịch trong PostgreSQL.
- Kinh nghiệm triển khai thực tế trên các nền tảng đám mây (Fly.io/Neon).

---

## 🚀 Limitations & Improvements
1. **Hiện tại còn thiếu gì?**
   - Chưa dùng ORM nên mã nguồn truy vấn trực tiếp hơi dài.
   - Chưa có bộ nhớ đệm (caching) cho lượt nhấp hoặc link.
   - ID chưa được làm xáo trộn (obfuscated).

2. **Nếu có thêm thời gian?**
   - Tích hợp OCR để đọc link từ hình ảnh hoặc mã QR.
   - Cho phép tùy chỉnh tên link ngắn (Custom alias).
   - Phân tích sâu hơn về thiết bị, vị trí IP của người nhấp.

3. **Để sẵn sàng cho môi trường Production?**
   - Cấu hình HTTPS và bảo mật header.
   - Thiết lập tự động mở rộng (Auto-scaling) và cân bằng tải (Load Balancer).
