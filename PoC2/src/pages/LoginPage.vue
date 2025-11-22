<template>
  <div class="page">
    <div class="card">
      <h1 class="title">KIKONECT</h1>
      <input
        type="email"
        placeholder="Email"
        class="input"
        v-model="email"
      />
      <input
        type="password"
        placeholder="Password"
        class="input"
        v-model="password"
      />
      <button class="login-btn" @click="handleLogin">Login</button>
    </div>
  </div>
</template>

<script setup>
import { ref } from "vue";
const email = ref("");
const password = ref("");
function handleLogin() {
  alert(
    `Login attempt:\nEmail: ${email.value}\nPassword: ${password.value}`
  );
}
</script>

<style scoped>
.page {
  width: 100vw;
  height: 100vh;
  background: #f9eefc;
  display: flex;
  justify-content: center;
  align-items: center;
}

.card {
  width: 300px;
  background: white;
  padding: 30px 25px;
  border-radius: 15px;
  box-shadow: 0 5px 20px rgba(0, 0, 0, 0.1);
  display: flex;
  flex-direction: column;
  gap: 20px;
  text-align: center;
}

.title {
  font-size: 22px;
  font-weight: 700;
  margin: 0;
}

.input {
  width: 90%;
  padding: 12px;
  border: 1px solid #cfcfcf;
  border-radius: 8px;
  font-size: 14px;
}

.login-btn {
  background: #e9ddfc;
  color: #6746ce;
  padding: 12px;
  border: none;
  border-radius: 20px;
  font-weight: 600;
  cursor: pointer;
  transition: 0.25s;
}

.login-btn:hover {
  opacity: 0.85;
}
</style>
