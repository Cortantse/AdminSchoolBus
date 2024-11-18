<!-- views/HomePage.vue -->
<template>
  <div>
    <h1>欢迎来到首页！</h1>
    <button @click="handleLogout">登出</button>
    <p v-if="message">{{ message }}</p>
  </div>
</template>

<script>

export default {
  name: "HomePage",
  data() {
    return {
      message: ""
    };
  },
  mounted() {
    const token = localStorage.getItem('jwtToken');
    if (!token) {
      alert('令牌无效或过期，请重新登录');
      this.$router.push('/login');
    }
  },
  methods: {
    handleLogout() {
      // 移除令牌并跳转到登录页
      localStorage.removeItem('jwtToken');
      alert('您已成功登出');
      this.$router.push('/login');
    },
  }
};
</script>

<style scoped>
button {
  margin-right: 10px;
}
</style>
