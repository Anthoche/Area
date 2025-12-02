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
    const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/;
    if (!emailRegex.test(email)) {
      alert("Please enter a valid email address.");
      return;
    }
    try {
      const res = await fetch("http://localhost:8080/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });
      if (res.status === 401 || res.status === 403) {
        alert("Invalid email or password");
        return;
      }
      if (!res.ok) {
        alert("Server error. Please try again later.");
        return;
      }
      const data = await res.json();
      console.log("Login success:", data);
      navigate("/home");
    } catch (err) {
      console.error("Network or fetch error:", err);
      alert("Network error.");
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
        onGithubLogin={() => alert('Github login clicked - to be implemented OAuth flow')}
        />
        }
      </div>
    </div>
  );
} 
