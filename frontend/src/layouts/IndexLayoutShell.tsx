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

export default function IndexLayoutShell({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <>
      <Theme theme="g100">
        <HeaderContainer
          render={({ isSideNavExpanded, onClickSideNavExpand }) => (
            <>
              <Header aria-label="Header">
                <SkipToContent />
                <HeaderMenuButton
                  aria-label={isSideNavExpanded ? 'Close menu' : 'Open menu'}
                  onClick={onClickSideNavExpand}
                  isActive={isSideNavExpanded}
                  aria-expanded={isSideNavExpanded}
                />
                <HeaderName href="/" prefix="KUURA" />
                {/* <HeaderName href="/" prefix="KUURA">
                  WELCOME
                </HeaderName> */}
              </Header>
            </>
          )}
        />
      </Theme>

      <Content>{children}</Content>
    </>
  );
}
