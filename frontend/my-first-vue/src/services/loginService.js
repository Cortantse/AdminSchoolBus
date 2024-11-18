// src/services/loginService.js
import axios from "axios";

export default {
    data() {
        return {
            username: "", // 邮箱账号
            password: "", // 密码
        };
    },
    methods: {
        /**
         * 处理用户的登录请求
         *
         * 此函数会验证用户输入的账号和密码是否为空，并向服务器发送登录请求。
         * 若成功，则保存令牌并进行后续处理；若失败，则给出错误信息提示。
         *
         * @returns {void} 无返回值
         * @throws {null} 如果请求失败或网络错误，在控制台和用户界面显示错误信息
         */
        handleLogin() {

            if (this.username === "" || this.password === "") {
                alert('账号或密码不能为空');
                return;
            }

            // 发送post请求
            axios.post("http://localhost:8888/api/login", {
                username: this.username,
                password: this.password,
            },
                {
                    timeout: 5000, // 设置超时时间，否则axios会一直等待
                })
                .then(response => {
                    if (response.data.code === 200) {
                        console.log("登陆成功：", response.data.data);
                        alert('登录成功');

                        // 保存令牌到 localStorage
                        localStorage.setItem('jwtToken', response.data.data);

                        // 跳转到首页
                        this.$router.push('/home');
                    } else {
                        alert('登录失败');
                        console.log("登陆失败：", response.data.message);
                    }
                })
                .catch(error => {
                    // 错误处理
                    if (error.response) {
                        console.log("服务器错误:", error.response.data);
                        alert("登录失败，服务器错误：" + (error.response.data.message || "未知错误"));
                    } else if (error.request) {
                        console.log("请求未收到响应:", error.request);
                        alert("登录失败，未收到服务器响应，请检查网络或服务器状态");
                    } else {
                        console.log("请求配置错误:", error.message);
                        alert("登录失败，错误信息：" + error.message);
                    }
                });
        },
    },
};
