import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import "./createacc.css";
import logo from "../../../lib/assets/Kikonect_logo.png";
import logoGoogle from "../../../lib/assets/G_logo.png";
import logoGithub from "../../../lib/assets/github_logo.png";

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

  const handleGoogleLogin = () => alert("Google login clicked");
  const handleGithubLogin = () => alert("Github login clicked");

  return (
    <div className="reg-page">
      <div className="reg-card">
        <img src={logo} alt="KiKoNect logo" className="logoR-img" />
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
