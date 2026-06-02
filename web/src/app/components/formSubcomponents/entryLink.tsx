import {
    ActionTypes,
    useRfidReaderContext
} from "@/app/components/formSubcomponents/readerWriterButtons/readerOptsContext";
import {ReactNode} from "react";
import {BaseExternalUrl} from "@/app/components/Constants";
import {Entry, EntryUrlId, viewUrlFor} from "@/app/components/common";

export default function EntryLink(
    {
        props,
    }: {
        props: {
            entryType: string,
            linkId: string,
            displayId?: string,
            openInNewTab?: boolean;
        };
    }) {
    return <EntryLinkWrapperForId props={{
        linkId: props.linkId,
        openInNewTab: props.openInNewTab || false, // TODO: false ok?
        entryType: props.entryType,
    }}><div  data-cy-id="EntryLink">
        {props.displayId || props.linkId}
    </div></EntryLinkWrapperForId>
}

export function EntryLinkWrapper<T extends Entry>(
    {
        props,
        children,
    }: {
        props: {
            entry: T;
            openInNewTab?: boolean;
        };
        children: ReactNode;
    }) {
    const onClickStopPropagation = (e: React.MouseEvent) => {
        e.stopPropagation();
    }
    const actualLink = viewUrlFor(props.entry.entryType(), EntryUrlId(props.entry))
    if (props.openInNewTab===true){
        return <a href={actualLink} target={"_blank"} rel={"noopener noreferrer"} onClick={onClickStopPropagation}>
            {children}
        </a>
    }
    return <a href={actualLink} onClick={onClickStopPropagation}>
        {children}
    </a>
}

export function EntryLinkWrapperForId(
    {
        props,
        children,
    }: {
        props: {
            linkId: string;
            entryType: string;
            openInNewTab?: boolean;
        };
        children: ReactNode;
    }) {
    const onClickStopPropagation = (e: React.MouseEvent) => {
        e.stopPropagation();
    }
    const actualLink = viewUrlFor(props.entryType, props.linkId)
    if (props.openInNewTab===true){
        return <a href={actualLink} target={"_blank"} rel={"noopener noreferrer"} onClick={onClickStopPropagation}>
            {children}
        </a>
    }
    return <a href={actualLink} onClick={onClickStopPropagation}>
        {children}
    </a>
}