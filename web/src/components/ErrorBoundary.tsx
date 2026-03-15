import { Component, type ReactNode } from "react";

import { ErrorMessage } from "./ErrorMessage";

type ErrorBoundaryProps = {
  children: ReactNode;
};

type ErrorBoundaryState = {
  hasError: boolean;
  message: string | null;
};

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  state: ErrorBoundaryState = {
    hasError: false,
    message: null
  };

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return {
      hasError: true,
      message: error.message || "A page-level error interrupted rendering."
    };
  }

  componentDidCatch(error: Error) {
    console.error("Page render failed", error);
  }

  render() {
    if (this.state.hasError) {
      return (
        <ErrorMessage
          title="Page crashed"
          message={this.state.message ?? "A page-level error interrupted rendering."}
        />
      );
    }

    return this.props.children;
  }
}
