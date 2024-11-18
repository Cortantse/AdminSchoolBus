<template>
  <div class="main">
    <div class="bg"></div>
    <div class="page">
    <h2>登录</h2>
    <input type="text" placeholder="用户名" class = "input" v-model="username"/>
    <input type="password" placeholder="密码" class = "input" v-model="password"/>
    <button class = "login-button" @click="handleLogin">登录</button>
    </div>
  </div>
</template>

<script>
import loginService from "../services/loginService.js";

export default {
  name: "LoginPage", // 修改为多单词名称
  data() {
    return loginService.data(); // 使用函数调用返回对象
  },
  methods: loginService.methods, // 使用外部的 methods
};
</script>


<style scoped>
.main {
  width: 99vw;
  height: 95vh;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bg {
  width: 100%;
  height: 100%;
  background: url("https://images.unsplash.com/photo-1728467459756-211f3c738697?q=80&w=2564&auto=format&fit=crop&ixlib=rb-4.0.3&ixid=M3wxMjA3fDB8MHxwaG90by1wYWdlfHx8fGVufDB8fHx8fA%3D%3D") no-repeat center center;
  background-size: cover;
  position: absolute;
  filter: blur(5px);
  z-index: -1;
}


.page {
  width: 300px;
  height: 300px;
  background-color: rgba(255, 255, 255, 0.6);
  border-radius: 8px;
  box-shadow: 0 10px 20px rgba(0, 0, 0, 0.5);
  padding: 30px;
  box-sizing: border-box;
  position: relative; /* 设置 position: relative 以配合 z-index */
  z-index: 1; /* 将 .page 提升到背景层之上 */
}

.input {
  width: 100%;
  padding: 10px;
  margin: 10px 0;
  border: 1px solid #ddd;
  border-radius: 4px;
  box-sizing: border-box;
}

.login-button {
  width: 100%;
  padding: 10px;
  background-color: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
  font-size: 16px;
  margin-top: 10px;
}

.login-button:hover {
  background-color: #0056b3;
}

</style>
