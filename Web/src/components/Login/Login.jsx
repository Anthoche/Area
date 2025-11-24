import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import LoginForm from "./LoginForm";
import "./login.css";
import logo from "../../../lib/assets/Kikonect_logo.png";

export default function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const handleForgotPassword = () => {
    alert("Forgot password clicked. Implement password reset flow.");
  };
  const handleSubmit = async (e) => {
  e.preventDefault();
  if (!email || !password) {
    alert("Please fill in all fields.");
    return;
  }

  const mockApiUrl = "https://6924b40b82b59600d721165a.mockapi.io/test";
  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!email || !password) {
      alert("Please fill in all fields.");
      return;
    }
    const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/;
    if (!emailRegex.test(email)) {
      alert("Please enter a valid email address.");
      return;
    }

    try {
      const res = await fetch(mockApiUrl);
      if (!res.ok) throw new Error("Failed to fetch users");
      const users = await res.json();
      const user = users.find(u => u.email === email && u.password === password);
      if (user) {
        alert("Login successful!");
        console.log("Logged in user:", user);
      } else {
        alert("Invalid email or password.");
      }
    } catch (err) {
      console.error("Network or server error:", err);
      alert("Network or server error. Please check your connection or try again.");
    }
  };
  return (
    <div className="login-page">
      <div className="login-card">
      <img src={logo} alt="KiKoNect logo" className="logo-img" />
        {
        <LoginForm
        email={email}
        setEmail={setEmail}
        password={password}
        setPassword={setPassword}
        handleSubmit={handleSubmit}
        handleForgotPassword={handleForgotPassword}
        onGoogleLogin={() => alert('Google login clicked - to be implemented OAuth flow')}
        onCreateAccount={() => alert('Github login clicked - to be implemented OAuth flow')}
        />
        }
      </div>
    </div>
    );
  }
}
