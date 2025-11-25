import React, { useState } from "react";
import "./register.css";
import logo from "../../../lib/assets/Kikonect_logo.png";

export default function Register() {
    const [firstName, setFirstName] = useState("");
    const [lastName, setLastName] = useState("");
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [confirm, setConfirm] = useState("");

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!email || !password || !confirm || !firstName || !lastName) {
            alert("Please fill in all fields.");
            return;
        }
        const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/;
        if (!emailRegex.test(email)) {
            alert("Please enter a valid email address.");
            return;
        }
        if (password !== confirm) {
            alert("Passwords do not match.");
            return;
        }
        try {
            const res = await fetch("http://localhost:8080/register", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify({
                    email,
                    password,
                    firstname: firstName,
                    lastname: lastName,
                }),
                }
            );
            if (res.status === 409) {
                alert("Email already registered. Please login instead.");
                return;
            }
            if (!res.ok) {
                alert("Server error. Please try again later.");
                return;
            }
            const data = await res.json();
            console.log("Registration success:", data);
            alert("Registration successful! You can now login.");
            setEmail("");
            setPassword("");
            setConfirm("");
            setFirstName("");
            setLastName("");
        } catch (err) {
            console.error("Network or fetch error:", err);
            alert("Network error. Please check your connection or backend.");
        }
    };

    return (
        <div className="reg-page">
        <div className="reg-card">
                <img src={logo} alt="KiKoNect logo" className="logo-img" />
            <h2 className="title">Create an account</h2>
            <form onSubmit={handleSubmit} className="reg-form">
            <input
                type="text"
                placeholder="First Name"
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                className="input-field"
            />
            <input
                type="text"
                placeholder="Last Name"
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                className="input-field"
            />
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
            <input
                type="password"
                placeholder="Confirm Password"
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                className="input-field"
            />
            <button type="submit" className="register-btn">
                Continue
            </button>
            </form>
        </div>
        </div>
    );
}
