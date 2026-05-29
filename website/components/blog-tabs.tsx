import Link from "next/link";

type Tab = { label: string; href: string; key: "release" | "engineering" };

const TABS: Tab[] = [
  { label: "Releases", href: "/blog", key: "release" },
  { label: "Engineering", href: "/blog/engineering", key: "engineering" },
];

export function BlogTabs({ active }: { active: "release" | "engineering" }) {
  return (
    <div className="mb-12">
      <nav className="blog-tabs">
        {TABS.map((tab) => (
          <Link
            key={tab.key}
            href={tab.href}
            aria-current={tab.key === active ? "page" : undefined}
            className={`blog-tab${tab.key === active ? " blog-tab-active" : ""}`}
          >
            {tab.label}
          </Link>
        ))}
      </nav>
    </div>
  );
}
