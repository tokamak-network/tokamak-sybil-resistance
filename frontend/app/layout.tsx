import type { Metadata } from "next";
import React from "react";
import "./globals.css";
import Header from "./components/Header";

export const metadata: Metadata = {
  title: "SYB",
  description: "Tokamak Sybil-Resistance",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <body className="flex flex-col min-h-screen">
        {/* Header */}
        <Header />
        {/* Main content area */}
        <main className="flex-grow container mx-auto p-4">{children}</main>
        {/* Footer or additional layout */}
        <footer className="bg-gray-800 text-white p-4 text-center">
          @ {`${new Date().getUTCFullYear()} SYB. All rights reserved.`}
        </footer>
      </body>
    </html>
  );
}
