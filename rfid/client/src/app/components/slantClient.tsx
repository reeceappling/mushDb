'use client'

import React, {JSX, useState} from "react";
import {IsValidNote, NewEntryNotes, Note, NotesFormArea} from "@/app/components/formSubcomponents/notes";
import {
    AllEntries,
    OnViewCreatorQuadCol,
    SplitAllEntries
} from "@/app/components/formSubcomponents/shared";
import ID from "@/app/components/formSubcomponents/id";
import DateArea from "@/app/components/formSubcomponents/date";
import {
    SlantData
} from "@/app/components/slantServer";
import {
    InitialPicsEntries, IsValidPicWithNotesIncoming,
    NewPicWithNotesForm,
    PicWithNotesForm,
} from "@/app/components/formSubcomponents/picWithNotes";
import ImageSelector from "@/app/components/formSubcomponents/imageSelector";
import {InnocDisplay, TransfersOutDisplay} from "@/app/components/transferClient";
import {KnownFruitableArea} from "@/app/components/formSubcomponents/knownFruitableArea";
import {GenerationInput} from "@/app/components/formSubcomponents/generationInput";
import {
    DisplayFormWrapper,
    DisplayInput, ExistingRecentSelector, FlexedArea, FlexedSinglesGroup,
    HandleJsonResponse,
    ImportDisplayInput, ImportEntryFormWrapper, importUrlFor,
    ListPageItems, ListPageTable, ListTableColumn, NewColumn, NewEntryFormWrapper,
    NewEntryInput, NumberToDateStr,
    OptionalArrayOfType, OptionalKey,
    OptionalSimpleKey,
    resolveContamsFormData,
    resolvePicsFormData, SendMultipartRequest, setFormData,
    setFormImages, viewUrlFor,
} from "@/app/components/common";
import ReaderWriterSelector, {
    WriteRfidOvcArea
} from "@/app/components/formSubcomponents/readerWriterButtons/readerSelector";
import {redirect} from "next/navigation";
import {
    ErrorDisplay,
    GensFormDisplay, MostRecentImageDisplay,
    ParentDisplay,
    PicsDisplay
} from "@/app/components/formSubcomponents/commonClient";
import {
    AgarBatchArea,
} from "@/app/components/agarBatchClient";
import {
    ContaminationForm, ContamsDisplay, InitialContamState, InitialNotesState, IsValidContamination,
    NewContaminationForm
} from "@/app/components/formSubcomponents/contaminations";
import {AgarBatchData, AgarBatchSelectorCloseable} from "@/app/components/agarBatchServer";
import {EntryLinkWrapper} from "@/app/components/formSubcomponents/entryLink";
import {BaseExternalUrl} from "@/app/components/Constants";
import {SpeciesData} from "@/app/components/speciesServer";
import {SubspeciesData} from "@/app/components/subspeciesServer";
import {SaleArea} from "@/app/components/saleClient";
import {ExistingSpeciesSelector, SpeciesSubspeciesArea} from "@/app/components/speciesClient";
import {ExistingSubSpeciesSelector} from "@/app/components/subspeciesClient";
import {AclDisplay, IsValidAcl, MarshalAcl, TogglableAreaWithDepth} from "@/app/components/accessControlClient";
import {ACL} from "@/app/components/accessControlServer";
import {OnViewCreatorsQuadColArea} from "@/app/components/formSubcomponents/ovc";
import {CreatedUpdatedDisposedArea} from "@/app/components/commonServer";

export function AssertSlant(input: any): asserts input is SlantData {
    if (typeof input !== 'object') {
        throw new Error('Input is not an object! Input is ' + typeof input);
    }
    // required simple keys
    let requiredSimpleKeys = new Map<string, string>([
        ['_id', 'string'],
        ['creationDate', 'number'],
        ['lastUpdated', 'number'],
    ])
    for (let [key, expType] of requiredSimpleKeys) {
        if (!(key in input && typeof input[key] === expType)) {
            throw new Error('Slant assertion failure: ' + key + 'was not type ' + expType + '. Was ' + (typeof input[key]));
        }
    }
    // optional simple keys
    let optionalSimpleKeys = new Map<string, string>([
        ['agarBatch', 'string'],
        ['stickType','string'],
        ['species', 'string'],
        ['subspecies', 'string'],
        ['innoc', 'string'],
        ['genSpore', 'number'],
        ['genFruitOrSpore', 'number'],
        ['parentType', 'string'],
        ['parent', 'string'],
        ['knownFruitable', 'boolean'],
        ['sale', 'string'],
        ['disposed', 'number'],
    ])
    for (let [key, expType] of optionalSimpleKeys) {
        if (!OptionalSimpleKey(key, input, expType)) {
            throw new Error('Slant assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional keys
    let complexOptionalKeys = new Map<string, (v: any) => boolean>([
        ['mostRecentImage', IsValidPicWithNotesIncoming],
       ['acl', IsValidAcl]
    ])
    for (let [key, validator] of complexOptionalKeys) {
        if (!OptionalKey(key, input, validator)) {
            throw new Error('Slant assertion failure: optional key ' + key + ' was not valid');
        }
    }
    // complex optional array keys
    let complexOptionalArrayKeys = new Map<string, (v: any) => boolean>([
        ['transfersOut', (item) => {
            return typeof item === 'string'
        }],
        ['pics', IsValidPicWithNotesIncoming],
        ['contamination', IsValidContamination],
        ['notes', IsValidNote],
    ])
    for (let [key, validator] of complexOptionalArrayKeys) {
        if (!OptionalArrayOfType(key, input, validator)) {
            throw new Error('Slant assertion failure: optional array key ' + key + ' was not valid');
        }
    }
    return
}

export function SlantImportDisplay({headerLevel, cookies}:ImportDisplayInput) {
    const [created, setCreated] = useState<number>(Date.now())
    const [stickType, setStickType] = useState<string | undefined>(undefined)
    const [species, setSpecies] = useState<SpeciesData | undefined>()
    const [subspecies, setSubspecies] = useState<SubspeciesData | undefined>()
    const [knownFruitable, setKnownFruitable] = useState<boolean | undefined>()
    const [generation, setGeneration] = useState<number | undefined>()
    const [imageFile, setImageFile] = useState<File | undefined>()
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>()
    const [err, setErr] = useState<string | undefined>()
    const ImportSlant = () => {
        let formData = new FormData()
        if(species===undefined){
            setErr("Species must be set!")
            return
        }
        formData.set('data', JSON.stringify({
            created:created,
            stickType: stickType,
            species: species._id,
            subspecies: subspecies?._id,
            knownFruitable: knownFruitable,
            generation: generation,
            writeTagTo: writeTagTo,
        }))
        if(imageFile!==undefined){
            formData.set("image", imageFile, "imgFile")
        }


        SendMultipartRequest(importUrlFor("slant"), cookies, formData)
            .then(HandleJsonResponse) // TODO: all of these for imports should be HandleJsonResponse, NOT HandleTxtResponse
            .then((newItem) => {
                AssertSlant(newItem)
                redirect(viewUrlFor("slant",newItem._id))
            })
            .catch((err) => {
                setErr(JSON.stringify(err))
            });
    }
    return <ImportEntryFormWrapper entryType={"slant"}>
        <ErrorDisplay err={err} headerLevel={headerLevel}/>
        <DateArea pre={"Created: "} when={created} readonly={false} updateParent={setCreated}/>
        <ExistingSpeciesSelector doSelect={setSpecies} headerLevel={headerLevel/*cookies={cookies}*/}/>
        {species?<ExistingSubSpeciesSelector species={species?._id} doSelect={setSubspecies} headerLevel={headerLevel/*cookies={cookies}*/}/>:null}
        <KnownFruitableArea doSelect={setKnownFruitable} headerLevel={headerLevel}/>
        <GenerationInput updateParent={setGeneration}/>
        <ImageSelector updateParent={setImageFile}/>
        <SlantStickSelector setStickType={setStickType}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo} headerLevel={headerLevel}/>
        <button className={"greenButton"} onClick={ImportSlant}>{"Import Slant"}</button>
    </ImportEntryFormWrapper>
}

export default function SlantDisplay(
    {
        id, readonly, data, headerLevel, isTopLevel, cookies
    }: DisplayInput) {
    try {
        AssertSlant(data)
        const [initial, setInitial] = useState(data)
        const [images, setImages] = useState<SplitAllEntries<PicWithNotesForm, NewPicWithNotesForm>>(InitialPicsEntries(initial.pics))
        const [contams, setContams] = useState<SplitAllEntries<ContaminationForm, NewContaminationForm>>(InitialContamState(initial.contamination))
        const [knownFruitable, setKnownFruitable] = useState(initial.knownFruitable)
        const [sale, setSale] = useState(initial.sale)
        const [disposed, setDisposed] = useState(initial.disposed)
        const [notes, setNotes] = useState<AllEntries<Note>>(InitialNotesState(data.notes))
        // Helper states
        const [transfersOut, setTransfersOut] = useState<string[]>(initial.transfersOut || [])
        const [err, setErr] = useState<string | undefined>()
        const [acl, setAcl] = useState<ACL | undefined>(initial.acl)
        const updateInitial = (updated: SlantData)=>{
            setInitial(updated)
            setImages(InitialPicsEntries(updated.pics))
            setContams(InitialContamState(updated.contamination))
            setKnownFruitable(updated.knownFruitable)
            setSale(updated.sale)
            setDisposed(updated.disposed)
            setNotes(InitialNotesState(updated.notes))
            // Helper states
            setTransfersOut(updated.transfersOut || [])
            setAcl(updated.acl)
        }
        const slantSubmit = ()=>{
            let body = new FormData()
            let dataObj:any={
                knownFruitable:knownFruitable,
                sale: sale,
                disposed:disposed,
                notes: notes,
                acl:MarshalAcl(acl),
            }
            try {
                // Pics
                let picsInfo = resolvePicsFormData(images)
                let newImages = picsInfo.images
                dataObj.images = picsInfo.obj
                // Contams
                let contamsInfo = resolveContamsFormData(contams)
                let newContams = contamsInfo.images
                dataObj.contams = contamsInfo.obj
                // Set data on form
                setFormData(body, dataObj)
                setFormImages(body, "newPic", newImages)
                setFormImages(body, "newContam", newContams)
            } catch (caught: any) {
                setErr(JSON.stringify(caught))
                return
            }

            SendMultipartRequest(BaseExternalUrl+"/db/update/slant/"+initial._id, cookies, body)
                .then(HandleJsonResponse)
                .then((entry) => {
                    AssertSlant(entry)
                    updateInitial(entry)
                    //window.location.reload()
                })
                .catch((er) => {
                    setErr(JSON.stringify(er))
                });
        }
        const ovcs: OnViewCreatorQuadCol[] = [
            WriteRfidOvcArea(initial._id),
        ]
        return (
            <DisplayFormWrapper entryType={"slant"}>
                <ErrorDisplay err={err} headerLevel={headerLevel}/>
                <ID id={data._id} txt={"Slant"} entryType={"slant"} />
                <OnViewCreatorsQuadColArea OnViewCreators={ovcs} readonly={readonly}/>
                <MostRecentImageDisplay data={initial.mostRecentImage} headerLevel={headerLevel}/>
                <FlexedArea>
                    <FlexedSinglesGroup>
                        <CreatedUpdatedDisposedArea created={initial.creationDate} updated={initial.lastUpdated} disposed={disposed}
                                                    readonly={readonly} setDisposedOnParent={setDisposed}/>
                    </FlexedSinglesGroup>
                    <FlexedSinglesGroup>
                        <SpeciesSubspeciesArea subspecies={initial.subspecies} species={initial.species}/>
                        <AgarBatchArea agarBatchId={initial.agarBatch} headerLevel={headerLevel}/>
                        <div>
                            {"Stick type:"+(initial.stickType || "none")}
                        </div>
                    </FlexedSinglesGroup>
                    {/* TODO: CondensationCoverageAtSealTimeField `bson:"inline"` // Percentage of condensation surface area coverage at seal time
                PourCoverageField                   `bson:"inline"` // Percentage of bottom surface area agar coverage
                WetAtCooledTimeField                `bson:"inline"` // Wet when initially cooled? True, false, or unknown
                AgarOnOutsideAtPourTimeFiel*/}
                    <FlexedSinglesGroup>
                        <InnocDisplay innoc={initial.innoc}/>
                        <ParentDisplay parent={initial.parent} parentType={initial.parentType} headerLevel={headerLevel}/>
                        <KnownFruitableArea initial={knownFruitable} doSelect={setKnownFruitable} readonly={readonly}/>
                        <SaleArea sale={sale} setSale={setSale} readonly={readonly} headerLevel={headerLevel}
                                  canCreateSale={true}/>
                    </FlexedSinglesGroup>
                    <GensFormDisplay gensSinceSpore={initial.genSpore} gensSinceFruitOrSpore={initial.genFruitOrSpore}/>
                </FlexedArea>

                <TransfersOutDisplay thisId={initial._id} thisEntryType={"slant"} transfersOut={transfersOut} allowNewTransferCreation={!readonly} cookies={cookies}/>
                <PicsDisplay pix={initial.pics || []} updateParent={setImages} readonly={readonly} headerLevel={headerLevel}/>{/* Pics */}
                <ContamsDisplay initial={initial.contamination || []} updateParent={setContams} readonly={readonly} headerLevel={headerLevel}/>
                <NotesFormArea readonly={readonly} initial={initial.notes} updateParent={setNotes}/>
                <TogglableAreaWithDepth startOpen={false} openTxt={"view permissions"} closeTxt={"minimize perms area"}>
                    <AclDisplay ACL={acl} readonly={readonly} updateParent={setAcl} />
                </TogglableAreaWithDepth>
                <DateArea pre={"Last Updated: "} when={initial.lastUpdated} readonly={true}/>
                {readonly ? null : <button className={"bottomButton greenButton"} onClick={(e)=>{
                    e.stopPropagation();
                    slantSubmit()
                }}>{"Update"}</button>}

            </DisplayFormWrapper>
        )
    } catch (err) {
        return <div>{"ERROR: Slant data format incorrect: " + err}</div>
    }
}

export function NewSlantForm({handlers,agarBatchIn}: {handlers: NewEntryInput<SlantData>, agarBatchIn?:AgarBatchData}){
    const [agarBatch, setAgarBatch] = useState(agarBatchIn)
    const [stickType, setStickType] = useState<string | undefined>(undefined)
    const [notes, setNotes] = useState<Note[]>([])
    const [writeTagTo, setWriteTagTo] = useState<string | undefined>(undefined)
    const [err, setErr] = useState<string | undefined>(undefined)
    const createSlant = (e: React.MouseEvent)=>{
        e.preventDefault()
        if(agarBatch===undefined){
            setErr("An agar batch must be selected")
            return
        }
        let body: any = {
            agarBatch: agarBatch,
            stickType: stickType,
            notes: notes,
            writeTagTo: writeTagTo,
        }
        fetch(BaseExternalUrl+"/db/create/slant", {
            method: "POST",
            headers: {
                credentials: 'include',
                'Content-type': "application/json"
            },
            body: JSON.stringify(body)
        })
            .then(HandleJsonResponse)
            .then((entry) => {
                AssertSlant(entry)
                handlers.onCreate && handlers.onCreate(entry)
            })
            .catch((error) => {
                setErr(JSON.stringify(error))
            });
    }
    return <NewEntryFormWrapper entryType={"slant"}>
        <ErrorDisplay err={err}/>
        <SlantStickSelector setStickType={setStickType}/>
        <AgarBatchSelectorCloseable doSelect={setAgarBatch} allowCreation={true} creatorInPage={true}/> {/* TODO: use new one instead?*/}
        <NewEntryNotes setNotes={setNotes}/>
        <ReaderWriterSelector txt={"Write to: "} defaultOption={"none"} onSelect={setWriteTagTo}/>
        <button className={"greenButton"} onClick={createSlant} onSubmit={(e)=>{e.preventDefault();}}>{"Create"}</button>
    </NewEntryFormWrapper>
}

// TODO: validate works properly
export function SlantStickSelector(x: { setStickType: (s?: string) => void }){
    return <div>
        {"Slant stick: "}<select className={"tailwindSelector"} defaultValue={"none"} disabled={false} onChange={(e) => {x.setStickType((e.currentTarget.value==="none")?undefined:e.currentTarget.value)}}>
        {["none","tongueDepressor"].map((s, i: number) => {
            return <option value={s} key={i}>{s}</option>
        })}
    </select>
    </div>
}

export function SlantListPageTable({data, onClick, withLink}: ListPageItems<SlantData>) {
    let cols: ListTableColumn<SlantData>[] = [
        NewColumn("ID", (v)=>v._id),
        NewColumn("Created", (v)=>{
            return NumberToDateStr(v.creationDate)
        }),
        NewColumn("Spec", (v)=>v.species||""),
        NewColumn("Subspec", v=>v.subspecies||"" ),
        NewColumn("Updated", (v)=>{
            return NumberToDateStr(v.lastUpdated)
        }),
    ]
    if (withLink) {
        cols = [...cols, NewColumn("Link", (v: SlantData)=>{
            return <EntryLinkWrapper props={{linkId:encodeURI(v._id),entryType:"slant",openInNewTab:true}}>
                <button className={"basicButtonSmall"}>{"View"}</button>
            </EntryLinkWrapper>
        })]
    }
    return <ListPageTable cols={cols} data={data} onClick={onClick}/>
}
export function SlantSelectorTable({data, onClick}: ListPageItems<SlantData>) {
    return <SlantListPageTable data={data} onClick={onClick} withLink={true} />
}
export function SlantSelector(
    {
        doSelect,
        allowCreate
    }: {
        doSelect: (val: SlantData | undefined) => void,
        allowCreate?: boolean
    }) {
    const table = (items: SlantData[]):JSX.Element=>{
        return <SlantSelectorTable data={items} onClick={doSelect}/>
    }

    return <ExistingRecentSelector entryType={"slant"} entryTypes={"slants"} doSelect={doSelect} asserter={AssertSlant}
                                   table={table}>
        {allowCreate && <NewSlantForm handlers={{onCreate: doSelect,isTopLevel: false}}/>}
    </ExistingRecentSelector>
}