/**
 * @file HowItWorks.jsx
 * @description
 * Section explaining how the automation platform works for users.
 *
 * Allows users to:
 *  - Learn how to build automations
 *  - Follow step-by-step onboarding
 */

"use client";
import "./welcomepage.css";
import "./hero-animations.css";
import { useEffect, useRef, useState } from "react";
import { Link } from "react-router-dom";

export default function HowItWorks() {
    const userId = Number(localStorage.getItem("user_id"));
    const isLoggedIn = Number.isFinite(userId) && userId > 0;
    const sectionRef = useRef(null);
    const [visible, setVisible] = useState(false);

    // Observe section visibility for animations
    useEffect(() => {
        const observer = new window.IntersectionObserver(
            ([entry]) => setVisible(entry.isIntersecting),
            { threshold: 0.2 }
        );
        if (sectionRef.current) 
            observer.observe(sectionRef.current);
        return () => observer.disconnect();
    }, []);
    return (
        <div ref={sectionRef} className={`welcome-page-section how-it-works hero-animate${visible ? ' visible' : ''}`} id="how-it-works">
            <h2 className="how-title">How it works</h2>
            <p className="how-subtitle">
               Create powerful automations in three simple steps
            </p>
            <div className="how-steps">
                <div className={`how-step hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.1s' : '0s' }}>
                    <div className="step-circle">1</div>
                    <h3>Choose your trigger</h3>
                    <p>Select the app and event that starts your automation. For example, "When I receive an email".</p>
                    <span className="step-example">Gmail, Dropbox, Twitter, Slack…</span>
                </div>
                <div className={`how-step hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.25s' : '0s' }}>
                    <div className="step-circle">2</div>
                    <h3>Add an action</h3>
                    <p>Choose what happens next. You can add multiple actions and create complex workflows with conditional logic.</p>
                    <span className="step-example">Save to Drive, Send message…</span>
                </div>
                <div className={`how-step hero-shape-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.4s' : '0s' }}>
                    <div className="step-circle">3</div>
                    <h3>Activate and relax</h3>
                    <p>Turn on your automation and let it run in the background. Monitor performance and tweak as needed.</p>
                    <span className="step-example">Running 24/7 automatically</span>
                </div>
            </div>
            <div className={`how-cta hero-animate${visible ? ' visible' : ''}`} style={{ transitionDelay: visible ? '0.6s' : '0s' }}>
                <h3>Ready to automate your workflow?</h3>
                <p>Join millions of users who save hours every week with automated workflows. Start for free, no credit card required.</p>
                <Link
                    className="how-cta-btn"
                    to={isLoggedIn ? "/home" : "/login"}
                >
                    Get Started for Free
                </Link>           
            </div>
        </div>
    );
}
