import React from "react";

export default function LoginForm( {
  email, setEmail, 
  password, setPassword, 
  handleSubmit, 
  handleForgotPassword,
  onGoogleLogin,
  onGithubLogin,
  onCreateAccount
}) {
  return (
    <form onSubmit={handleSubmit}>
      <input
      type="email"
      placeholder="Email"
      value={email}
      onChange={e => setEmail(e.target.value)}
      className="input-field"
      />
      <input
      type="password"
      placeholder="Passeword"
      value={password}
      onChange={e => setPassword(e.target.value)}
      className="input-field"
      />
      <div className="forgot-row">
        <button type="button" className="forgot-btn" onClick={handleForgotPassword}>
          Forgot password?
        </button>
      </div>
      <button type="submit" className="login-btn">
        Login
      </button>
      <div className="social-login-raw">
        <button type="button" className="social-btn google" onClick={onGoogleLogin}>
          Login with Google
        </button>
        <button type="button" className="social-btn github" onClick={onGithubLogin}>
          Login with Github
        </button>
      </div>
      <div className="create-account-row">
        <button type="button" className="forgot-btn" onClick={onCreateAccount}>
          Don’t have an account ? Sign In
        </button>
      </div>
    </form>
  );
}
