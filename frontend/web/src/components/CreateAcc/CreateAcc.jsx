import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./createacc.css";
import logo from "../../../lib/assets/Kikonect_logo_no_text.png";
import logoGoogle from "../../../lib/assets/G_logo.png";
import logoGithub from "../../../lib/assets/github_logo.png";

const API_BASE =
  import.meta.env.VITE_API_URL ||
  import.meta.env.API_URL ||
  `${window.location.protocol}//${window.location.hostname}:8080`;

export default function CreateAcc() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = (e) => {
    e.preventDefault();
    setError("");
    const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/;
    if (!emailRegex.test(email)) {
      setError("Invalid email");
      return;
    }
    navigate("/register", { state: { email } });
  };

  const handleGoogleLogin = () => {
    const id = Number(localStorage.getItem("user_id"));
    const uiRedirect = encodeURIComponent(`${window.location.origin}/home`);
    const baseUrl = `${API_BASE}/oauth/google/login?ui_redirect=${uiRedirect}`;
    const url = id && id > 0 ? `${baseUrl}&user_id=${id}` : baseUrl;
    window.location.href = url;
  };
  const handleGithubLogin = () => {
    const id = Number(localStorage.getItem("user_id"));
    const uiRedirect = encodeURIComponent(`${window.location.origin}/home`);
    const baseUrl = `${API_BASE}/oauth/github/login?ui_redirect=${uiRedirect}`;
    const url = id && id > 0 ? `${baseUrl}&user_id=${id}` : baseUrl;
    window.location.href = url;
  };

  return (
    <div className="reg-page">
      <div className="reg-card">
          <div className="logo-container">
              <img src={logo} alt="KiKoNect logo" className="logo-img" />
              <h1>KiKoNect</h1>
          </div>
        <h2 className="title">Create an account</h2>
        <p className="sub-text">Enter your email to sign up for this app</p>
        {error && (
          <div className="error-popup">
            <strong>Error:</strong> {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="reg-form">
          <div className="floating-input">
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
            />
            <label className={email ? "filled" : ""}>Email</label>
          </div>
          <button type="submit" className="reg-btn">
            Continue
          </button>
        </form>
        <div className="or-separator">
          <span>or</span>
        </div>
        <div className="social-login-raw">
          <button
            type="button"
            className="login-btn google"
            onClick={handleGoogleLogin}
          >
            Sign in with Google
            <img src={logoGoogle} alt="G_logo" className="logoG-img" />
          </button>
          <button
            type="button"
            className="login-btn github"
            onClick={handleGithubLogin}
          >
            Sign in with Github
            <img src={logoGithub} alt="github_logo" className="logoG-img" />
          </button>
        </div>
      </div>
    </div>
  );
}
