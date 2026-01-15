/**
 * @file Register.jsx
 * @description
 * User registration form for new account creation.
 *
 * Allows users to:
 *  -  Handle form validation
 *  -  Handle password rules
 *  -  Create an account via backend API
 */

import React, { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import logo from "../../../lib/assets/Kikonect_logo_no_text.png";
import eyeOpen from "../../../lib/assets/eye_open.png";
import eyeClosed from "../../../lib/assets/eye_closed.png";
import "./register.css";

/**
 * Resolve backend API base URL.
 */
const API_BASE =
    import.meta.env.VITE_API_URL ||
    import.meta.env.API_URL ||
    `${window.location.protocol}//${window.location.hostname}:8080`;

export default function Register() {
    const location = useLocation();
    const navigate = useNavigate();

    // Prefill email if coming from previous step (CreateAcc)
    const prefilledEmail = location.state?.email || "";
    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [email, setEmail] = useState(prefilledEmail);
    const [password, setPassword] = useState("");
    const [confirm, setConfirm] = useState("");
    const [formError, setFormError] = useState("");
        const [showPassword, setShowPassword] = React.useState(false);
            const [showConfirmPassword, setShowConfirmPassword] = React.useState(false);


    useEffect(() => {
        if (prefilledEmail) setEmail(prefilledEmail);
    }, [prefilledEmail]);

    /**
     * Handles registration submission.
     * Performs:
     *  - Empty field validation
     *  - Email format validation
     *  - Password complexity & match validation
     *  - API call to backend
     */
    const handleSubmit = async (e) => {
        e.preventDefault();
        setFormError("");
        if (!email || !password || !confirm || !firstName || !lastName) {
            setFormError("Please fill in all fields.");
            return;
        }
        const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/;
        if (!emailRegex.test(email)) {
            setFormError("Please enter a valid email address.");
            return;
        }
        // Password rules: min 8 chars, at least one number, one special char
        const passwordRules = /^(?=.*[0-9])(?=.*[!@#$%^&*()_+\-=[\]{};':"\\|,.<>/?]).{8,}$/;
        if (!passwordRules.test(password)) {
            setFormError("Password must be at least 8 characters, include a number and a special character.");
            return;
        }
        if (password !== confirm) {
            setFormError("Passwords do not match.");
            return;
        }

        try {
            const res = await fetch(`${API_BASE}/register`, {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    email,
                    password,
                    firstname: firstName,
                    lastname: lastName,
                }),
            });

            if (res.status === 409) {
                setFormError("Email already registered. Please login instead.");
                return;
            }

            if (!res.ok) {
                setFormError("Server error. Please try again later.");
                return;
            }

            const data = await res.json();
            console.log("Registration success:", data);
            setFormError("");
            setEmail("");
            setPassword("");
            setConfirm("");
            setFirstName("");
            setLastName("");
            navigate("/login");
        } catch (err) {
            console.error("Network or fetch error:", err);
            setFormError("Network error. Please check your connection or backend.");
        }
    };

    return (
        <div className="reg-page">
            <div className="reg-card">
                <div className="logo-container">
                    <img src={logo} alt="KiKoNect logo" className="logo-img" />
                    <h1>KiKoNect</h1>
                </div>
                <h2 className="title">Create an account</h2>

                {formError && (
                    <div
                        className="error-popup"
                        style={{
                            marginBottom: 30,
                            color: "#b91818ff",
                            fontStyle: "italic",
                            fontWeight: 300,
                            fontSize: 15,
                        }}
                    >
                        {formError}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="reg-form">
                    <div className="floating-input">
                        <input
                            type="text"
                            value={firstName}
                            onChange={(e) => setFirstName(e.target.value)}
                            required
                        />
                        <label className={firstName ? "filled" : ""}>First Name</label>
                    </div>
                    <div className="floating-input">
                        <input
                            type="text"
                            value={lastName}
                            onChange={(e) => setLastName(e.target.value)}
                            required
                        />
                        <label className={lastName ? "filled" : ""}>Last Name</label>
                    </div>
                    <div className="floating-input">
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            required
                        />
                        <label className={email ? "filled" : ""}>Email</label>
                    </div>
                    <div className="floating-input password-wrapper">
                        <input
                            type={showPassword ? "text" :"password"}
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                        />
                        <label className={password ? "filled" : ""}>
                            Password
                        </label>
                        <button
                            type="button"
                            className="toggle-password"
                            onClick={() => setShowPassword(!showPassword)}
                            tabIndex={-1}
                        >
                            <img
                                src={showPassword ? eyeClosed : eyeOpen}
                                alt={showPassword ? "Hide password" : "Show password"}
                                className="eye-img"
                            />
                        </button>
                    </div>
                    <div className="floating-input password-wrapper">
                        <input
                            type={showConfirmPassword ? "text" :"password"}
                            value={confirm}
                            onChange={(e) => setConfirm(e.target.value)}
                            required
                        />
                        <label className={confirm ? "filled" : ""}>
                            Confirm Password
                        </label>
                        <button
                            type="button"
                            className="toggle-password"
                            onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                            tabIndex={-1}
                        >
                            <img
                                src={showConfirmPassword ? eyeClosed : eyeOpen}
                                alt={showConfirmPassword ? "Hide password" : "Show password"}
                                className="eye-img"
                            />
                        </button>
                    </div>
                    <button type="submit" className="reg-btn">
                        Register
                    </button>
                </form>
            </div>
        </div>
    );
}
