'use client'

import React, {JSX, useContext, useState} from "react";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    CreatedLinkFor,
    DisplayFormWrapper,
    DisplayInput,
    DoCreateRequestMultipart,
    DoMultipartImportRequest,
    DoUpdateMultipartRequest,
    ExistingRecentSelector,
    FlexedArea,
    FlexedSinglesGroup,
    ImportDisplayInput,
    ImportEntryFormWrapper,
    ListPageItems,
    ListPageTable,
    ListTableColumn,
    NewColumn,
    NewEntryFormWrapper,
    NumberToDateStr,
    OptionalArrayOfType,
    OptionalKey,
    OptionalSimpleKey,
    RequiredKey,
    resolvePicsFormData,
    setFormFull
} from "@/app/components/common";
import {
    DisposedDisplay,
    ErrorDisplay,
    MostRecentImageDisplay,
    PicsDisplay,
    SporePrintColorArea,
    SporePrintDensityArea,
} from "@/app/components/formSubcomponents/commonClient";
import {
    InitialPicsEntries,
    IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import ID from "@/app/components/formSubcomponents/id";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {SpeciesData} from "@/app/components/speciesServer";
import {SporePrintData} from "@/app/components/sporePrintServer";
import {SaleArea} from "@/app/components/saleClient";
import {
    AddCreatedQuadColFunction,
    AllEntries,
    Data,
    OnViewCreatorQuadCol
} from "@/app/components/formSubcomponents/shared";
import EntryLinkForId, {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {ExistingSpeciesSubspeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ChildMssArea, NewMssForm} from "@/app/components/mssClient";
import {FruitSelectorCloseable} from "@/app/components/fruitServer";
import {AclDisplay, MarshalAcl, TogglableAreaWithDepth, UnmarshalAcl} from "@/app/components/accessControlClient";
import {ChildSwabArea, NewSporeSwabForm} from "@/app/components/sporeSwabClient";
import {ACL} from "@/app/components/accessControlServer";
import {MssData} from "@/app/components/mssServer";
import {InitialNotesState} from "@/app/components/formSubcomponents/initialState";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {allCookies, CookiesContext} from "@/app/components/formSubcomponents/cookiesContext/cookies";
import {ConfirmOrCancel} from "@/app/components/formSubcomponents/moveOnceUsed";
import {ActionTypes, useModalContext} from "@/app/components/formSubcomponents/modalContext/modal";

export function AssertSporePrint(input: any): asserts input is SporePrintData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }

    // required simple keys
    const requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['species', 'string'],
        ['lastUpdated', 'number'],
    ])
    for (const [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Plate assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    const optionalSimpleKeys = new Map<string, string>([
        ['parent', 'string'],
        ['subspecies', 'string'],
        ['sale', 'string'],
        ['disposed', 'number'],
        ['color', 'string'],
        ['density', 'string'],
    ])
    for (const [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Plate assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
    ])
    for (const [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Spore Print assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex required keys
    const complexRequiredKeys = new Map<string, (v: any) => boolean>([
        //['acl', IsValidAcl]
    ])
    for (const [key, validator] of complexRequiredKeys) {
        if (!RequiredKey(key, input, validator)) {
            throw new Error('Spore Print assertion failure: required key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    const complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['pics', IsValidPicWithNotesIncoming],
        ['notes', IsValidNote],
    ])
    for (const [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Plate assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    // Unmarshal ACL
    if (!('acl' in input)) {
        throw 'ACL missing from input in asserter'
    }
    input.acl = UnmarshalAcl(input.acl)
    return
}

export default function SporePrintDisplay(
    {
        readonly, data, headerLevel, isTopLevel
    }: DisplayInput<SporePrintData>) {
    const {dispatch} = useModalContext();
    const [initial, setInitial] = useState(data)
    const initNotes: Data<Note>[] = (data.notes || []).map((n) => {
        return {data: n, disabled: false}
    })
    const [color, setColor] = useState<string | undefined>()
    const [density, setDensity] = useState<string | undefined>()
    const [pics, setPics] = useState(InitialPicsEntries(data.pics))
    const [sale, setSale] = useState(data.sale)
    const [disposed, setDisposed] = useState(data.disposed)
    const [notes, setNotes] = useState<AllEntries<Note>>({existing: initNotes, new: []})
    const [err, setErr] = useState<string | undefined>()
    const [acl, setAcl] = useState<ACL>(data.acl)
    const updateInitial = (updated: SporePrintData) => {
        setInitial(updated)
        setColor(updated.color)
        setDensity(updated.density)
        setPics(InitialPicsEntries(updated.pics))
        setSale(updated.sale)
        setDisposed(updated.disposed)
        setNotes(InitialNotesState(updated.notes))
        setAcl(updated.acl)
        setErr(undefined)
    }
    const cookies = useContext(CookiesContext)
    const submit = () => {
        // sale disposed, project, pics, notes
        const formData = new FormData()
        const dataObj: any = {
            // All optional but acl
            color: color,
            density: density,
            sale: sale,
            disposed: disposed,
            notes: notes,
            acl: MarshalAcl(acl),
        }
        try {
            // Pics
            const picsInfo = resolvePicsFormData(pics)
            const newImages = picsInfo.images
            dataObj.images = picsInfo.obj
            // Set data on form
            setFormFull(formData, dataObj, newImages, undefined, undefined)
            // formData.set("data", JSON.stringify(dataObj))
            // setFormImages("newPic", formData, newImages)
        } catch (caught: any) {
            setErr(JSON.stringify(caught))
            return
        }

        DoUpdateMultipartRequest("sporePrint", data._id, formData, AssertSporePrint, allCookies(cookies))
            .then(v => {
                updateInitial(new SporePrintData(v))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Success",
                        text: "entry updated successfully",
                        isErr: false
                    }})
            })
            .catch(e => {
                setErr("failed to update initial: " + JSON.stringify(e))
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Update Failed",
                        text: "failed to update: " + JSON.stringify(e),
                        isErr: true
                    }})
            })
    }
    const Nowdate = new Date()
    const innoculated = true // Spore prints are always considered innoculated
    const ovcs: () => OnViewCreatorQuadCol[] = () => {
        const disp = initial.disposed !== undefined
        const daysCutoff = 10 // TODO: ensure cutoff ok
        const createdAtLeastNDaysAgo = Nowdate.getTime() - initial.creationDate > (1000 * 60 * 60 * 24 * daysCutoff)
        return [
            // TODO: test heavily for all
            // OVC for chaining prints
            ...(innoculated && !createdAtLeastNDaysAgo && initial.parent!==undefined?[
                {
                    txt: "Create another print from parent", // TODO: validate works properly!!! Created on 8/7/26
                    newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                        return <NewSporePrintForm onCreate={(item: SporePrintData) => {
                            onCreate([{
                                typeText: "Spore Print",
                                node: <CreatedLinkFor linkId={item._id} typ={"sporePrint"}/>,
                            }], true)
                        }} parentId={initial._id}/>
                    },
                }
            ]:[]),
            ...(innoculated && !disp ?[
                {
                    txt: "Create Spore Swab",
                    newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                        return <NewSporeSwabForm printIn={data} onCreate={(item: MssData) => { // TODO: switch to handlers{{}} format
                            onCreate([{
                                typeText: "Spore Swab",
                                node: <CreatedLinkFor linkId={item._id} typ={"sporeSwab"}/>,
                            }], false)
                        }}/>
                    },
                },
                {
                    txt: "Create MultiSpore Syringe",
                    newCreationArea: (onCreate: AddCreatedQuadColFunction) => {
                        return <NewMssForm sporePrintIn={data} handlers={{
                            isTopLevel: false,
                            onCreate: (item: MssData) => {
                                onCreate([{
                                    typeText: "Multispore Syringe",
                                    node: <CreatedLinkFor linkId={item._id} typ={"mss"}/>,
                                }], false)
                            }
                        }}/>
                    },
                },
                // TODO: print transfer to agar?
                // TODO: TRANSFERS SKIPPING SWABS/SYRINGES?! Probably not...
            ]:[]),
            // OVCs that always exist
            WriteRfidOvcArea(initial._id),
        ]
    }

    return <DisplayFormWrapper entryType={"sporePrint"}>
        <ErrorDisplay err={err}/>
        <ID props={{id: data._id, txt: "Spore Print", entryType: "sporePrint"}}/>
        <OnViewCreatorsQuadColArea OnViewCreators={ovcs()} readonly={readonly}/>
        <MostRecentImageDisplay data={initial.mostRecentImage}/>
        <FlexedArea>
            <FlexedSinglesGroup>
                <DateArea pre={"Print Date: "} readonly={true} when={data.creationDate}/>
                <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                <DisposedDisplay readonly={false} initial={initial.disposed} setDisposedOnParent={setDisposed}/>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <SpeciesSubspeciesArea species={initial.species} subspecies={initial.subspecies}/>
                <div>
                    <div>{"Parent: "}</div>
                    {initial.parent ?
                        <EntryLinkForId props={{
                            displayId: initial.parent,
                            linkId: initial.parent,
                            entryType: "fruit",
                            openInNewTab: true
                        }}/>
                        : "Store"}
                </div>
            </FlexedSinglesGroup>
            <FlexedSinglesGroup>
                <SporePrintColorArea readonly={readonly || initial.color !== undefined || initial.disposed !== undefined}
                                     color={initial.color} setColor={setColor}/>
                <SporePrintDensityArea readonly={readonly || initial.density !== undefined || initial.disposed !== undefined}
                                       density={initial.density} setDensity={setDensity}/>
                <SaleArea readonly={readonly} canCreateSale={true} sale={sale} setSale={setSale}/>
            </FlexedSinglesGroup>
        </FlexedArea>
        <ChildMssArea parent={initial._id}/>{/* TODO: area where we can display all the child MSS of this print? */}
        <ChildSwabArea parent={initial._id}/>{/* TODO: area where we can display all the child swabs of this print? */}
        <PicsDisplay pix={initial.pics || []} updateParent={setPics} readonly={readonly}
                     headerLevel={headerLevel}/>{/* Pics */}
        <NotesFormArea initial={initial.notes} readonly={readonly} updateParent={setNotes}/>
        <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
            <AclDisplay initial={initial.acl} readonly={readonly} updateParent={setAcl}/>
        </TogglableAreaWithDepth>
        {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e) => {
            e.stopPropagation();
            submit()
        }}>{"Update"}</button>}
    </DisplayFormWrapper>
}

// Make spore prints directly from fruit, or indirectly from many others!
export function NewSporePrintForm(
    {parentId, headerLevel, offset, onCreate}: {
        parentId?: string
        headerLevel?: number
        offset?: number
        onCreate: (sp: SporePrintData) => void
    }) {
    const {dispatch} = useModalContext();
    const [parent, setParent] = useState<string | undefined>(parentId)
    const [pics, setPics] = useState<NewPicWithNotesForm[]>([])
    const [notes, setNotes] = useState<Note[]>([])
    //const [parentType, setParentType] = useState<string | undefined>(parentTypeIn)
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)

    const cookies = useContext(CookiesContext)
    const createEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!parent) {
            setErr("Fruit must be selected")
            return
        }
        const doReq = () => {
            const formData = new FormData()
            const dataObj: any = {
                // parentType: parentType, // TODO: may be deletable!
                parent: parent,
                notes: notes,
                // optional pics also here
                writeTagTo: writeTagTo,
            }
            // Pics
            dataObj.pics = pics.map(p => {
                return {
                    time: p.time, notes: p.notes.new.map(n => {
                        return n.data
                    })
                }
            })
            // Perms
            formData.set("data", JSON.stringify(dataObj))
            for (let i = 0; i < pics.length; i++) {
                const toSend = pics[i]
                if (toSend.img === undefined) {
                    setErr("new image " + i + " is undefined")
                    return
                }
                const fileName = "newPic" + "-" + i
                formData.set(fileName, toSend.img, fileName)
            }
            DoCreateRequestMultipart("sporePrint", formData, AssertSporePrint, allCookies(cookies))
                .then(v => {
                    onCreate ? onCreate(v) : console.log("no onCreate provided")
                    dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Creation succeeded",
                            text: "created",
                            isErr: false
                        }})
                })
                .catch(e => {
                    const newErr = JSON.stringify(e)
                    setErr(newErr)
                    dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                            header: "Creation failed",
                            text: newErr,
                            isErr: true
                        }})
                })
        }
        // if both pics and notes are empty, do nothing
        if (pics.length === 0 && notes.length === 0) {
            ConfirmOrCancel({txt: "No notes or pictures, did you mean to do that?", onConfirm: doReq})
        } else {
            doReq()
        }
    }

    return <NewEntryFormWrapper entryType={"sporePrint"}>
        <ErrorDisplay err={err}/>
        {parentId === undefined && <FruitSelectorCloseable onSelect={(f) => {
            setParent(f?._id)
        }} hideDisposed={true}/>}
        <PicsDisplay pix={[]} readonly={false} updateParent={(ps) => {
            setPics(ps.new)
        }} headerLevel={headerLevel} offset={offset}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createEntry}>{"Create"}</button>
    </NewEntryFormWrapper>
}

export function SporePrintImportDisplay({headerLevel}: ImportDisplayInput) { // TODO: USE ONLY FOR EXISTING SPORE PRINTS!
    const {dispatch} = useModalContext();
    const [printDate, setPrintDate] = useState<number>(Date.now())
    const [color, setColor] = useState<string | undefined>()
    const [density, setDensity] = useState<string | undefined>()
    const [notes, setNotes] = useState<Note[]>([])
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<string | undefined>()
    const [image, setImage] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const cookies = useContext(CookiesContext)
    const importEntry = (e: React.MouseEvent) => {
        e.preventDefault()
        if (!species) {
            setErr("A species must be selected")
            return
        }
        const formData = new FormData()
        const dataObj: any = {
            creationDate: printDate,
            color: color,
            density: density,
            species: species._id,
            // optional
            subspecies: subspecies,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        formData.set("data", JSON.stringify(dataObj))
        if (image !== undefined) {
            formData.set("img", image, "img")
        }
        const dispatchUpdate = (isErr:boolean, text:string)=>{
            if(isErr){
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Creation failed",
                        text: text,
                        isErr: true
                    }})
            } else {
                dispatch({type: ActionTypes.SET_MODAL_INFO, payload:{
                        header: "Creation successful",
                        text: text,
                        isErr: false
                    }})
            }
        }
        DoMultipartImportRequest(formData, "sporePrint", AssertSporePrint, setErr, allCookies(cookies), dispatchUpdate)
    }
    //no parent because we couldn't possibly know it
    return <ImportEntryFormWrapper entryType={"sporePrint"}>
        <ErrorDisplay err={err}/>
        <DateArea pre={"Print Date: "} readonly={false} when={Date.now()} updateParent={setPrintDate}/>
        {/* TODO: parent! store or online?*/}
        <SporePrintColorArea readonly={false} setColor={setColor}/>
        <SporePrintDensityArea readonly={false} setDensity={setDensity}/>
        <ExistingSpeciesSubspeciesSelector doSelectSpecies={setSpecies} doSelectSubspecies={setSubspecies}/>
        <ImageSelector updateParent={setImage}/>
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={importEntry}>{"Create"}</button>
    </ImportEntryFormWrapper>
}

export function SporePrintListPageTable({data, onClick, withLink}: ListPageItems<SporePrintData>) {
    let cols: ListTableColumn<SporePrintData>[] = [
        NewColumn("ID", (v) => v._id, true),
        NewColumn("Created", (v) => {
            return NumberToDateStr(v.creationDate)
        }, true),
        NewColumn("Spec", (v) => v.species || "", true),
        NewColumn("Subspec", v => v.subspecies || "", true),
        NewColumn("Updated", (v) => {
            return NumberToDateStr(v.lastUpdated)
        }), // TODO: fit?
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SporePrintData) => {
            return <EntryLinkWrapper props={{entry: v, openInNewTab: true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick} newClass={v => {
        return new SporePrintData(v)
    }}/>
}

export function SporePrintSelectorTable({data, onClick}: ListPageItems<SporePrintData>) {
    return <SporePrintListPageTable data={data} onClick={onClick} withLink={true}/>
}

export function SporePrintSelector(
    {
        doSelect,
        hideDisposed = false
    }: {
        doSelect: (val: SporePrintData | undefined) => void,
        hideDisposed?: boolean
    }) {
    const table = (items: SporePrintData[]): JSX.Element => {
        return <SporePrintSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"sporePrint"} entryTypes={"sporePrints"} doSelect={doSelect}
                                   asserter={AssertSporePrint}
                                   table={table} hideDisposed={hideDisposed}>
    </ExistingRecentSelector>
}
