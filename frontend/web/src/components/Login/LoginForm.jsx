import React from "react";
import { useNavigate } from "react-router-dom";
import logoGoogle from "../../../lib/assets/G_logo.png";
import logoGithub from "../../../lib/assets/github_logo.png";

export default function LoginForm({
  email,
  setEmail,
  password,
  setPassword,
  handleSubmit,
  handleForgotPassword,
  onGoogleLogin,
  onGithubLogin
}) {

  const navigate = useNavigate();

  const goToRegister = () => {
    navigate("/register");
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="email"
        placeholder="Email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        className="input-field"
      />

      <input
        type="password"
        placeholder="Password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        className="input-field"
      />

      <div className="forgot-row">
        <button
          type="button"
          className="forgot-btn"
          onClick={handleForgotPassword}
        >
          Forgot password?
        </button>
      </div>

      <button type="submit" className="login-btn">
        Login
      </button>

      <div className="or-separator">
        <span>or</span>
      </div>

      <div className="social-login-raw">
        <button
          type="button"
          className="login-btn google"
          onClick={onGoogleLogin}
        >
          Login with Google
          <img src={logoGoogle} alt="G_logo" className="logoG-img" />

        </button>
        

        <button
          type="button"
          className="login-btn github"
          onClick={onGithubLogin}
        >
          Login with Github
          <img src={logoGithub} alt="github_logo" className="logoG-img" />
        </button>
      </div>

      <div className="create-account-row">
        <button
          type="button"
          className="forgot-btn"
          onClick={goToRegister}
        >
          Don’t have an account ? Sign Up
        </button>
      </div>
    </form>
  );
}
