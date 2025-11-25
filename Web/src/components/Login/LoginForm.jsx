import React from "react";
import { useNavigate } from "react-router-dom";

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
        Continue
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
          Continue with Google
        </button>

        <button
          type="button"
          className="login-btn github"
          onClick={onGithubLogin}
        >
          Continue with Github
        </button>
      </div>

      <div className="create-account-row">
        <button
          type="button"
          className="forgot-btn"
          onClick={goToRegister}
        >
          Don’t have an account ? Sign In
        </button>
      </div>
    </form>
  );
}
