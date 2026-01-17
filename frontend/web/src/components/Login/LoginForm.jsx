/**
 * @file LoginForm.jsx
 * @description
 * Login form UI for user authentication and OAuth options.
 *
 * Allows users to:
 *  - Sign up using email
 *  - Continue registration flow
 *  - Authenticate using Google or GitHub OAuth
 */

import React from "react";
import { useNavigate } from "react-router-dom";
import logoGoogle from "../../../lib/assets/G_logo.png";
import logoGithub from "../../../lib/assets/github_logo.png";
import eyeOpen from "../../../lib/assets/eye_open.png";
import eyeClosed from "../../../lib/assets/eye_closed.png";

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
        navigate("/createacc");
    };

    const [showPassword, setShowPassword] = React.useState(false);

    return (
        <form onSubmit={handleSubmit}>
            <div className="floating-input">
                <input
                    id="login-email"
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    autoComplete="email"
                    placeholder="Email"
                    required
                />
                <label htmlFor="login-email" className="sr-only">
                    Email
                </label>
            </div>

            <div className="floating-input password-wrapper">
                <input
                    id="login-password"
                    type={showPassword ? "text" :"password"}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    autoComplete="current-password"
                    placeholder="Password"
                    required
                />
                <label htmlFor="login-password" className="sr-only">
                    Password
                </label>
                <button
                    type="button"
                    className="toggle-password"
                    aria-label={showPassword ? "Hide password" : "Show password"}
                    aria-pressed={showPassword}
                    onClick={() => setShowPassword(!showPassword)}
                >
                    <img
                        src={showPassword ? eyeClosed : eyeOpen}
                        alt={showPassword ? "Hide password" : "Show password"}
                        className="eye-img"
                    />
                </button>
            </div>

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
                    <img
                        src={logoGoogle}
                        alt="G_logo"
                        className="logoG-img"
                    />
                </button>

                <button
                    type="button"
                    className="login-btn github"
                    onClick={onGithubLogin}
                >
                    Login with Github
                    <img
                        src={logoGithub}
                        alt="github_logo"
                        className="logoG-img"
                    />
                </button>
            </div>

            <div className="create-account-row">
                <button
                    type="button"
                    className="forgot-btn"
                    onClick={goToRegister}
                >
                    Don’t have an account? Sign Up
                </button>
            </div>
        </form>
    );
}
