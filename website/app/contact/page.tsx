"use client";

import { useState, FormEvent } from "react";
import { ChevronDown } from "lucide-react";
import { Eyebrow } from "@/components/eyebrow";

const COUNTRIES = [
  "United States",
  "Canada",
  "United Kingdom",
  "Germany",
  "France",
  "Netherlands",
  "Spain",
  "Italy",
  "Sweden",
  "Poland",
  "India",
  "Australia",
  "Japan",
  "Singapore",
  "Brazil",
  "Mexico",
  "Other",
];

const COMPANY_SIZES = [
  "Just me",
  "2 – 10",
  "11 – 50",
  "51 – 200",
  "201 – 500",
  "500+",
];

const CALENDLY_URL =
  process.env.NEXT_PUBLIC_CALENDLY_URL ??
  "https://calendly.com/tracewayapp/demo";

export default function Contact() {
  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [email, setEmail] = useState("");
  const [country, setCountry] = useState("");
  const [companySize, setCompanySize] = useState("");
  const [submitting, setSubmitting] = useState(false);

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setSubmitting(true);

    const params = new URLSearchParams();
    const fullName = `${firstName.trim()} ${lastName.trim()}`.trim();
    if (fullName) params.set("name", fullName);
    if (email.trim()) params.set("email", email.trim());
    if (country) params.set("a1", country);
    if (companySize) params.set("a2", companySize);

    const url = `${CALENDLY_URL}${
      params.toString() ? `?${params.toString()}` : ""
    }`;
    window.location.href = url;
  }

  const inputStyle: React.CSSProperties = {
    background: "color-mix(in oklab, var(--ink-0) 60%, transparent)",
    border: "1px solid var(--hair-2)",
    color: "var(--fg-0)",
    fontFamily: "var(--font-mono)",
  };

  const labelStyle: React.CSSProperties = {
    color: "var(--fg-0)",
    fontFamily: "var(--font-display)",
  };

  const helperStyle: React.CSSProperties = {
    color: "var(--fg-3)",
    fontFamily: "var(--font-mono)",
  };

  return (
    <main className="relative">
      <section className="hero gridbg relative overflow-hidden pt-20 pb-24">
        <div className="wrap relative z-10 max-w-3xl mx-auto">
          <div className="text-center mb-10">
            <Eyebrow>Book a demo</Eyebrow>
            <h1 className="mt-4">
              We&apos;d love to hear from you.{" "}
              <em>Submit the form to pick a time.</em>
            </h1>
            <p
              className="mt-5 text-[17px] max-w-[560px] mx-auto"
              style={{ color: "var(--fg-2)" }}
            >
              Tell us a bit about you and your team. On submit we&apos;ll hand
              you over to Calendly so you can grab a 30-minute slot.
            </p>
          </div>

          <form
            onSubmit={handleSubmit}
            className="rounded-[14px] p-6 md:p-8"
            style={{
              background:
                "linear-gradient(180deg, color-mix(in oklab, var(--ink-2) 80%, transparent), color-mix(in oklab, var(--ink-1) 60%, transparent))",
              border: "1px solid var(--hair-2)",
              boxShadow: "0 30px 60px -30px rgba(0, 0, 0, 0.5)",
            }}
          >
            {/* Name row */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <input
                type="text"
                required
                value={firstName}
                onChange={(e) => setFirstName(e.target.value)}
                placeholder="First name"
                aria-label="First name"
                className="w-full px-4 py-3.5 rounded-md text-[15px] outline-none focus:border-[color:var(--a1)] transition-colors"
                style={inputStyle}
              />
              <input
                type="text"
                required
                value={lastName}
                onChange={(e) => setLastName(e.target.value)}
                placeholder="Last name"
                aria-label="Last name"
                className="w-full px-4 py-3.5 rounded-md text-[15px] outline-none focus:border-[color:var(--a1)] transition-colors"
                style={inputStyle}
              />
            </div>

            {/* Email */}
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="Email*"
              aria-label="Email address"
              className="mt-4 w-full px-4 py-3.5 rounded-md text-[15px] outline-none focus:border-[color:var(--a1)] transition-colors"
              style={inputStyle}
            />

            {/* Country */}
            <div className="mt-8">
              <label
                htmlFor="country"
                className="block text-[15px] font-medium"
                style={labelStyle}
              >
                Which country is your company&apos;s headquarters based in?{" "}
                <span style={{ color: "var(--crit)" }}>*</span>
              </label>
              <p className="mt-1 text-[13px]" style={helperStyle}>
                This info will help us connect you to the right person.
              </p>
              <div className="mt-3 relative">
                <select
                  id="country"
                  required
                  value={country}
                  onChange={(e) => setCountry(e.target.value)}
                  className="w-full appearance-none px-4 py-3.5 pr-10 rounded-md text-[15px] outline-none focus:border-[color:var(--a1)] transition-colors"
                  style={inputStyle}
                >
                  <option value="" disabled>
                    Select an option
                  </option>
                  {COUNTRIES.map((c) => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
                <ChevronDown
                  className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none"
                  size={18}
                  style={{ color: "var(--fg-3)" }}
                />
              </div>
            </div>

            {/* Company size */}
            <div className="mt-7">
              <label
                htmlFor="company-size"
                className="block text-[15px] font-medium"
                style={labelStyle}
              >
                How many employees work at your company?{" "}
                <span style={{ color: "var(--crit)" }}>*</span>
              </label>
              <p className="mt-1 text-[13px]" style={helperStyle}>
                This info will help us connect you to the right person.
              </p>
              <div className="mt-3 relative">
                <select
                  id="company-size"
                  required
                  value={companySize}
                  onChange={(e) => setCompanySize(e.target.value)}
                  className="w-full appearance-none px-4 py-3.5 pr-10 rounded-md text-[15px] outline-none focus:border-[color:var(--a1)] transition-colors"
                  style={inputStyle}
                >
                  <option value="" disabled>
                    Select an option
                  </option>
                  {COMPANY_SIZES.map((c) => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
                <ChevronDown
                  className="absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none"
                  size={18}
                  style={{ color: "var(--fg-3)" }}
                />
              </div>
            </div>

            <div className="mt-8 flex justify-center">
              <button
                type="submit"
                disabled={submitting}
                className="btn btn-accent px-10 py-3 min-w-[200px] justify-center disabled:opacity-60 disabled:cursor-not-allowed"
              >
                {submitting ? "Redirecting…" : "Submit"}
              </button>
            </div>

            <p
              className="mt-5 text-center text-[12px]"
              style={{
                color: "var(--fg-3)",
                fontFamily: "var(--font-mono)",
              }}
            >
              We&apos;ll redirect you to Calendly with your info pre-filled.
            </p>
          </form>
        </div>
      </section>
    </main>
  );
}
