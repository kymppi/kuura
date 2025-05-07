import {
  Content,
  Header,
  HeaderContainer,
  HeaderMenuButton,
  HeaderName,
  SkipToContent,
  Theme,
} from '@carbon/react';
import type React from 'react';

export default function LoginLayout({
  children,
}: {
  readonly children: React.ReactNode;
}) {
  return (
    <>
      <Theme theme="g100">
        <HeaderContainer
          render={({ isSideNavExpanded, onClickSideNavExpand }) => (
            <Header aria-label="Header">
              <SkipToContent />
              <HeaderMenuButton
                aria-label={isSideNavExpanded ? 'Close menu' : 'Open menu'}
                onClick={onClickSideNavExpand}
                isActive={isSideNavExpanded}
                aria-expanded={isSideNavExpanded}
              />
              <HeaderName href="/" prefix="KUURA">
                User login
              </HeaderName>
            </Header>
          )}
        />
      </Theme>

      <Content style={{ height: 'calc(100dvh - 48px)', padding: 0 }}>
        {children}
      </Content>
    </>
  );
}
