/**
 * @file Login.jsx
 * @description
 * Login page container for authentication logic and state.
 *
 * Allows users to:
 *  - Sign up using email
 *  - Continue registration flow
 *  - Authenticate using Google or GitHub OAuth
 */

import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import LoginForm from "./LoginForm";
import "./login.css";
import logo from "../../../lib/assets/Kikonect_logo_no_text.png";

/**
 * Resolve backend API base URL.
 * Falls back to current host on port 8080 for local development.
 */
const API_BASE =
    import.meta.env.VITE_API_URL ||
    import.meta.env.API_URL ||
    `${window.location.protocol}//${window.location.hostname}:8080`;

export default function Login() {
    const navigate = useNavigate();
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");

    /**
     * Initialize userId from localStorage if present.
     * Used to link OAuth accounts when logging in again.
     */
    const [userId, setUserId] = useState(() => {
        const stored = localStorage.getItem("user_id");
        return stored ? Number(stored) : null;
    });

    const handleForgotPassword = () => {
        alert("Forgot password clicked. Implement password reset flow.");
    };

    /**
     * Handles email/password login.
     * Performs basic validation and stores user data on success.
     */
    const handleSubmit = async (e) => {
        e.preventDefault();

        if (!email || !password) {
            alert("Please fill in all fields.");
            return;
        }

        // Basic email format validation
        const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/;
        if (!emailRegex.test(email)) {
            alert("Please enter a valid email address.");
            return;
        }

        try {
            const res = await fetch(`${API_BASE}/login`, {
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

            // Persist user identity for session continuity
            if (data?.id) {
                setUserId(data.id);
                localStorage.setItem("user_id", data.id);
            }

            if (data?.email) {
                localStorage.setItem("user_email", data.email);
            }

            navigate("/home");
        } catch (err) {
            console.error("Network or fetch error:", err);
            alert("Network error.");
        }
    };

    return (
        <div className="login-page">
            <div className="login-card">
                <div className="logo-container">
                    <img src={logo} alt="KiKoNect logo" className="logo-img" />
                    <h1>KiKoNect</h1>
                </div>

                <LoginForm
                    email={email}
                    setEmail={setEmail}
                    password={password}
                    setPassword={setPassword}
                    handleSubmit={handleSubmit}
                    handleForgotPassword={handleForgotPassword}
                    onGoogleLogin={() => {
                        /**
                         * Redirects to Google OAuth.
                         * Attaches user_id when available to link accounts.
                         */
                        const id = userId || Number(localStorage.getItem("user_id"));
                        const uiRedirect = encodeURIComponent(`${window.location.origin}/home`);
                        const baseUrl = `${API_BASE}/oauth/google/login?ui_redirect=${uiRedirect}`;
                        const url =
                            id && id > 0
                                ? `${baseUrl}&user_id=${id}`
                                : baseUrl;

                        window.location.href = url;
                    }}
                    onGithubLogin={() => {
                        /**
                         * Same OAuth flow as Google, but for GitHub.
                         */
                        const id = userId || Number(localStorage.getItem("user_id"));
                        const uiRedirect = encodeURIComponent(`${window.location.origin}/home`);
                        const baseUrl = `${API_BASE}/oauth/github/login?ui_redirect=${uiRedirect}`;
                        const url =
                            id && id > 0
                                ? `${baseUrl}&user_id=${id}`
                                : baseUrl;

                        window.location.href = url;
                    }}
                />
            </div>
        </div>
    );
}
