// Mock WorkOS components for Storybook
// This file mocks the WorkOS AuthKit components that require server-side functionality

import React from "react";

// Mock AuthKitProvider - just passes through children
export function AuthKitProvider({ children }: { children: React.ReactNode }) {
  return <>{children}</>;
}

// Mock user for Storybook stories
export const mockUser = {
  id: "user_mock123",
  email: "demo@example.com",
  firstName: "Demo",
  lastName: "User",
  profilePictureUrl: null,
};

// Mock useAuth hook
export function useAuth() {
  return {
    user: mockUser,
    isLoading: false,
    isAuthenticated: true,
  };
}

// Mock signIn function
export async function signIn() {
  console.log("Mock signIn called");
}

// Mock signOut function
export async function signOut() {
  console.log("Mock signOut called");
}
