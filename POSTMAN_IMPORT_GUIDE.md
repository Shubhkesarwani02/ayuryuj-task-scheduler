# 📬 POSTMAN IMPORT & TESTING GUIDE

## 🚀 **QUICK START**

### **Step 1: Import Postman Collection**
1. Open Postman
2. Click **Import** button (top left)
3. Select **File** tab
4. Choose: `docs/task-scheduler-api-complete.postman_collection.json`
5. Click **Import**

### **Step 2: Import Environment**
1. Click **Import** again
2. Choose one of:
   - `docs/local-environment.postman_environment.json` (for local development)
   - `docs/docker-environment.postman_environment.json` (for Docker)
   - `docs/production-environment.postman_environment.json` (for production)
3. Click **Import**
4. Select the imported environment (top right dropdown)

### **Step 3: Start Your Server**
```powershell
# Start the application locally
go run ./cmd/server/main.go

# Or with Docker:
docker-compose up --build
```

### **Step 4: Run Tests**
1. Open **Task Scheduler API - Complete Test Suite** collection
2. Click **Run Collection** button
3. Select all requests or specific folders
4. Click **Run**

## 📚 **COMPREHENSIVE DOCUMENTATION**

For detailed testing instructions, advanced configuration, and troubleshooting:
**👉 [docs/POSTMAN_TESTING_GUIDE.md](docs/POSTMAN_TESTING_GUIDE.md)**

---

## 📋 **COLLECTION OVERVIEW**

### **🔍 System Health & Info**
- `GET /health` - Service health check with component status
- `GET /api/v1/metrics` - System metrics and statistics

### **📋 Task Management - CRUD Operations**
- `POST /api/v1/tasks` - Create new task (one-off or cron)
- `GET /api/v1/tasks` - List all tasks with pagination
- `GET /api/v1/tasks/{id}` - Get specific task details
- `PUT /api/v1/tasks/{id}` - Update existing task
- `DELETE /api/v1/tasks/{id}` - Delete task

### **⚡ Task Execution & Control**
- `POST /api/v1/tasks/{id}/execute` - Execute task immediately
- `POST /api/v1/tasks/{id}/pause` - Pause recurring task
- `POST /api/v1/tasks/{id}/resume` - Resume paused task

### **📊 Results & History**
- `GET /api/v1/results` - All execution results with pagination
- `GET /api/v1/tasks/{id}/results` - Results for specific task

### **🧪 Advanced Test Scenarios**

### **📚 Documentation**
- `GET /swagger/index.html` - API documentation

---

## 🧪 **AUTOMATED TESTS INCLUDED**

### **✅ Response Validation**
- Status code checks (200, 201, 204)
- Response time validation (< 5000ms)
- Content-Type header verification

### **✅ Data Validation**
- JSON schema validation
- Required field presence
- UUID format validation
- Timestamp format checks

### **✅ Business Logic Tests**
- Task creation with different trigger types
- Proper CRUD operation sequencing
- Error handling validation

---

## 🔧 **ENVIRONMENT VARIABLES**

### **Local Development Environment**
```json
{
  "baseUrl": "http://localhost:8080/api/v1",
  "host": "localhost:8080",
  "taskId": "{{taskId}}",
  "testWebhookUrl": "https://httpbin.org/post",
  "testGetUrl": "https://httpbin.org/get"
}
```

### **Docker Environment** (if using docker-compose)
```json
{
  "baseUrl": "http://localhost:8080/api/v1",
  "host": "localhost:8080",
  "taskId": "{{taskId}}",
  "testWebhookUrl": "https://httpbin.org/post",
  "testGetUrl": "https://httpbin.org/get"
}
```

---

## 🎯 **EXPECTED TEST RESULTS**

When you run the collection, you should see:
- ✅ **25+ tests passing**
- ✅ **All requests returning 200/201 status**
- ✅ **No failed assertions**
- ✅ **Response times under 5 seconds**

---

## 🔧 **TROUBLESHOOTING**

### **Server Not Running**
```
Error: connect ECONNREFUSED 127.0.0.1:8080
```
**Solution**: Start the task scheduler application first

### **Database Not Connected**
```
Error: database connection failed
```
**Solution**: Ensure PostgreSQL is running (check docker-compose)

### **Port Already in Use**
```
Error: listen tcp :8080: bind: address already in use
```
**Solution**: Stop other applications using port 8080 or change port in environment

---

## 🎊 **READY TO TEST!**

Your Postman collection is comprehensive and ready for:
- ✅ **Development Testing**
- ✅ **API Validation**
- ✅ **Integration Testing**
- ✅ **Performance Monitoring**

**Happy Testing! 🚀**