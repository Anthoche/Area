import React, { useState, useEffect } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import "./register.css";
import logo from "../../../lib/assets/Kikonect_logo.png";

const API_BASE =
    import.meta.env.VITE_API_URL ||
    import.meta.env.API_URL ||
    `${window.location.protocol}//${window.location.hostname}:8080`;

export default function Register() {
    const location = useLocation();
    const navigate = useNavigate();
    const prefilledEmail = location.state?.email || "";

    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [email, setEmail] = useState(prefilledEmail);
    const [password, setPassword] = useState("");
    const [confirm, setConfirm] = useState("");
    const [formError, setFormError] = useState("");

    useEffect(() => {
        if (prefilledEmail) setEmail(prefilledEmail);
    }, [prefilledEmail]);

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
            navigate("/");
        } catch (err) {
            console.error("Network or fetch error:", err);
            setFormError("Network error. Please check your connection or backend.");
        }
    };

    return (
        <div className="reg-page">
            <div className="reg-card">
                <img src={logo} alt="KiKoNect logo" className="logoR-img" />
                <h2 className="title">Create an account</h2>
                                {formError && (
                                    <div
                                        className="error-popup"
                                        style={{ marginBottom: 30, color: "#b91818ff", fontStyle: 'italic', fontWeight: 300, fontSize: 15 }}
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
                    <div className="floating-input">
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            required
                        />
                        <label className={password ? "filled" : ""}>Password</label>
                    </div>
                    <div className="floating-input">
                        <input
                            type="password"
                            value={confirm}
                            onChange={(e) => setConfirm(e.target.value)}
                            required
                        />
                        <label className={confirm ? "filled" : ""}>Confirm Password</label>
                    </div>
                    <button type="submit" className="reg-btn">
                        Register
                    </button>
                </form>
            </div>
        </div>
    );
}