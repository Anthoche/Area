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
  
  const handleSubmit = (e) => {
    e.preventDefault();
    alert(`Login attempt:\nEmail: ${email}\nPassword: ${password}`);
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
