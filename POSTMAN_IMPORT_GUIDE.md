# 📬 POSTMAN IMPORT & TESTING GUIDE

## 🚀 **QUICK START**

### **Step 1: Import Postman Collection**
1. Open Postman
2. Click **Import** button (top left)
3. Select **File** tab
4. Choose: `docs/comprehensive_postman_collection.json`
5. Click **Import**

### **Step 2: Import Environment**
1. Click **Import** again
2. Choose: `docs/postman_environment_local.json`
3. Click **Import**
4. Select **Local Development** environment (top right dropdown)

### **Step 3: Start Your Server**
```powershell
# Start the application
.\task-scheduler.exe

# Or with Go:
go run ./cmd/server/main.go
```

### **Step 4: Run Tests**
1. Open **Task Scheduler API** collection
2. Click **Run Collection** button
3. Select all requests
4. Click **Run Task Scheduler API**

---

## 📋 **COLLECTION OVERVIEW**

### **🏥 Health & System**
- `GET /health` - Service health check
- `GET /api/v1/metrics` - System metrics

### **📋 Task Management**
- `POST /api/v1/tasks` - Create new task
- `GET /api/v1/tasks` - List all tasks
- `GET /api/v1/tasks/{id}` - Get specific task
- `PUT /api/v1/tasks/{id}` - Update task
- `DELETE /api/v1/tasks/{id}` - Delete task

### **📊 Results & Analytics**
- `GET /api/v1/results` - Task execution results

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