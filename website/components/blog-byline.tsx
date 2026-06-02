import Image from "next/image";

const AUTHOR = {
  name: "Dusan Stanojevic",
  url: "https://www.linkedin.com/in/dusanstanojeviccs",
  avatar: "/images/dusan-stanojevic.jpg",
};

export function BlogByline({ date }: { date: string }) {
  return (
    <div className="blog-byline">
      <a
        href={AUTHOR.url}
        target="_blank"
        rel="noopener noreferrer"
        className="blog-author"
      >
        <Image
          src={AUTHOR.avatar}
          alt={AUTHOR.name}
          width={32}
          height={32}
          className="blog-author-avatar"
        />
        {AUTHOR.name}
      </a>
      {date && (
        <>
          <span className="blog-byline-sep">·</span>
          <span>{date}</span>
        </>
      )}
    </div>
  );
}
